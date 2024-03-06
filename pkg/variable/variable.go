/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package variable

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	cgcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	kubekeyv1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1"
	kubekeyv1alpha1 "github.com/kubesphere/kubekey/v4/pkg/apis/kubekey/v1alpha1"
	_const "github.com/kubesphere/kubekey/v4/pkg/const"
	"github.com/kubesphere/kubekey/v4/pkg/variable/source"
)

type Variable interface {
	Key() string
	Get(option GetOption) (any, error)
	Merge(option ...MergeOption) error
}

type Options struct {
	Ctx      context.Context
	Client   ctrlclient.Client
	Pipeline kubekeyv1.Pipeline
}

// New variable. generate value from config args. and render to source.
func New(o Options) (Variable, error) {
	// new source
	s, err := source.New(RuntimeDirFromPipeline(o.Pipeline))
	if err != nil {
		klog.V(4).ErrorS(err, "create file source failed", "path", filepath.Join(RuntimeDirFromPipeline(o.Pipeline)), "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	// get config
	var config = &kubekeyv1.Config{}
	if err := o.Client.Get(o.Ctx, types.NamespacedName{o.Pipeline.Spec.ConfigRef.Namespace, o.Pipeline.Spec.ConfigRef.Name}, config); err != nil {
		klog.V(4).ErrorS(err, "get config from pipeline error", "config", o.Pipeline.Spec.ConfigRef, "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	// get inventory
	var inventory = &kubekeyv1.Inventory{}
	if err := o.Client.Get(o.Ctx, types.NamespacedName{o.Pipeline.Spec.InventoryRef.Namespace, o.Pipeline.Spec.InventoryRef.Name}, inventory); err != nil {
		klog.V(4).ErrorS(err, "get inventory from pipeline error", "inventory", o.Pipeline.Spec.InventoryRef, "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	v := &variable{
		key:    string(o.Pipeline.UID),
		source: s,
		value: &value{
			Config:    *config,
			Inventory: *inventory,
			Hosts: map[string]host{
				_const.LocalHostName: {}, // set default host
			},
		},
	}
	// read data from source
	data, err := v.source.Read()
	if err != nil {
		klog.V(4).ErrorS(err, "read data from source error", "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	for k, d := range data {
		if k == _const.RuntimePipelineVariableLocationFile {
			// set location
			if err := json.Unmarshal(d, &v.value.Location); err != nil {
				klog.V(4).ErrorS(err, "unmarshal location error", "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
				return nil, err
			}
		} else {
			// set hosts
			h := host{}
			if err := json.Unmarshal(d, &h); err != nil {
				klog.V(4).ErrorS(err, "unmarshal host error", "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
				return nil, err
			}
			v.value.Hosts[strings.TrimSuffix(k, ".json")] = h
		}
	}
	return v, nil
}

type GetOption interface {
	filter(data value) (any, error)
}

// KeyPath get a key path variable
type KeyPath struct {
	// HostName which host obtain the variable
	HostName string
	// LocationUID locate which variable belong to
	LocationUID string
	// Path base top variable.
	Path []string
}

func (k KeyPath) filter(data value) (any, error) {
	// find value from location
	var getLocationFunc func(uid string) any
	getLocationFunc = func(uid string) any {
		if loc := findLocation(data.Location, uid); loc != nil {
			// find value from task
			if v, ok := data.Hosts[k.HostName].RuntimeVars[uid]; ok {
				if result := k.getValue(v, k.Path...); result != nil {
					return result
				}
			}
			if result := k.getValue(loc.Vars, k.Path...); result != nil {
				return result
			}
			if loc.PUID != "" {
				return getLocationFunc(loc.PUID)
			}
		}
		return nil
	}
	if result := getLocationFunc(k.LocationUID); result != nil {
		return result, nil
	}

	// find value from host
	if result := k.getValue(data.Hosts[k.HostName].Vars, k.Path...); result != nil {
		return result, nil
	}

	// find value from global
	if result := k.getValue(data.getGlobalVars(k.HostName), k.Path...); result != nil {
		return result, nil
	}
	return nil, nil
}

// getValue from variable.VariableData use key path. if key path is empty return nil
func (k KeyPath) getValue(value VariableData, key ...string) any {
	if len(key) == 0 {
		return nil
	}
	var result any
	result = value
	for _, s := range key {
		result = result.(VariableData)[s]
	}
	return result
}

// ParentLocation UID for current location
type ParentLocation struct {
	LocationUID string
}

func (p ParentLocation) filter(data value) (any, error) {
	loc := findLocation(data.Location, p.LocationUID)
	if loc != nil {
		return loc.PUID, nil
	}
	return nil, fmt.Errorf("cannot find location %s", p.LocationUID)
}

// LocationVars get all variable for location
type LocationVars struct {
	// HostName which host obtain the variable
	// if HostName is empty. get value from global
	HostName string
	// LocationUID locate which variable belong to
	LocationUID string
}

func (b LocationVars) filter(data value) (any, error) {
	var result VariableData
	if b.HostName != "" {
		// find from host runtime
		if v, ok := data.Hosts[b.HostName].RuntimeVars[b.LocationUID]; ok {
			result = mergeVariables(result, v)
		}

		// merge location variable
		var mergeLocationVarsFunc func(uid string)
		mergeLocationVarsFunc = func(uid string) {
			// find value from task
			if v, ok := data.Hosts[b.HostName].RuntimeVars[uid]; ok {
				result = mergeVariables(result, v)
			}
			if loc := findLocation(data.Location, uid); loc != nil {
				result = mergeVariables(result, loc.Vars)
				if loc.PUID != "" {
					mergeLocationVarsFunc(loc.PUID)
				}
			}
		}
		mergeLocationVarsFunc(b.LocationUID)

		// get value from host
		result = mergeVariables(result, data.Hosts[b.HostName].Vars)
	}

	// get value from global
	result = mergeVariables(result, data.getGlobalVars(b.HostName))

	return result, nil
}

// HostVars get all top variable for a host
type HostVars struct {
	HostName string
}

func (k HostVars) filter(data value) (any, error) {
	return mergeVariables(data.getGlobalVars(k.HostName), data.Hosts[k.HostName].Vars), nil
}

// Hostnames from  array contains group name or host name
type Hostnames struct {
	Name []string
}

func (g Hostnames) filter(data value) (any, error) {
	var hs []string
	for _, n := range g.Name {
		// add host to hs
		if _, ok := data.Hosts[n]; ok {
			hs = append(hs, n)
		}
		// add group's host to gs
		for gn, gv := range convertGroup(data.Inventory) {
			if gn == n {
				hs = mergeSlice(hs, gv.([]string))
				break
			}
		}

		// Add the specified host in the specified group to the hs.
		regex := regexp.MustCompile(`^(.*)\[\d\]$`)
		if match := regex.FindStringSubmatch(n); match != nil {
			index, err := strconv.Atoi(match[2])
			if err != nil {
				klog.V(4).ErrorS(err, "convert index to int error", "index", match[2])
				return nil, err
			}
			for gn, gv := range data.Inventory.Spec.Groups {
				if gn == match[1] {
					hs = append(hs, gv.Hosts[index])
					break
				}
			}
		}
	}
	return hs, nil
}

type DependencyTasks struct {
	LocationUID string
}

type DependencyTask struct {
	Tasks    []string
	Strategy func([]kubekeyv1alpha1.Task) kubekeyv1alpha1.TaskPhase
}

func (f DependencyTasks) filter(data value) (any, error) {
	loc := findLocation(data.Location, f.LocationUID)
	if loc == nil {
		return nil, fmt.Errorf("cannot found location %s", f.LocationUID)

	}
	return f.getDependencyLocationUIDS(data, loc)
}

func (f DependencyTasks) getDependencyLocationUIDS(data value, loc *location) (DependencyTask, error) {
	if loc.PUID == "" {
		return DependencyTask{
			Strategy: func([]kubekeyv1alpha1.Task) kubekeyv1alpha1.TaskPhase {
				return kubekeyv1alpha1.TaskPhaseRunning
			},
		}, nil
	}

	// if tasks has failed. execute current task.
	failedExecuteStrategy := func(tasks []kubekeyv1alpha1.Task) kubekeyv1alpha1.TaskPhase {
		if len(tasks) == 0 { // non-dependency
			return kubekeyv1alpha1.TaskPhaseRunning
		}
		skip := true
		for _, t := range tasks {
			if !t.IsComplete() {
				return kubekeyv1alpha1.TaskPhasePending
			}
			if t.IsFailed() {
				return kubekeyv1alpha1.TaskPhaseRunning
			}
			if !t.IsSkipped() {
				skip = false
			}
		}
		if skip {
			return kubekeyv1alpha1.TaskPhaseRunning
		}
		return kubekeyv1alpha1.TaskPhaseSkipped
	}

	// If dependency tasks has failed. skip it.
	succeedExecuteStrategy := func(tasks []kubekeyv1alpha1.Task) kubekeyv1alpha1.TaskPhase {
		if len(tasks) == 0 { // non-dependency
			return kubekeyv1alpha1.TaskPhaseRunning
		}
		skip := true
		for _, t := range tasks {
			if !t.IsComplete() {
				return kubekeyv1alpha1.TaskPhasePending
			}
			if t.IsFailed() {
				return kubekeyv1alpha1.TaskPhaseSkipped
			}
			if !t.IsSkipped() {
				skip = false
			}
		}
		if skip {
			return kubekeyv1alpha1.TaskPhaseSkipped
		}
		return kubekeyv1alpha1.TaskPhaseRunning
	}

	// If dependency tasks is not complete. waiting.
	// If dependency tasks is skipped. skip.
	alwaysExecuteStrategy := func(tasks []kubekeyv1alpha1.Task) kubekeyv1alpha1.TaskPhase {
		if len(tasks) == 0 { // non-dependency
			return kubekeyv1alpha1.TaskPhaseRunning
		}
		skip := true
		for _, t := range tasks {
			if !t.IsComplete() {
				return kubekeyv1alpha1.TaskPhasePending
			}
			if !t.IsSkipped() {
				skip = false
			}
		}
		if skip {
			return kubekeyv1alpha1.TaskPhaseSkipped
		}
		return kubekeyv1alpha1.TaskPhaseRunning
	}

	// Find the parent location and, based on where the current location is within the parent location, retrieve the dependent tasks.
	ploc := findLocation(data.Location, loc.PUID)

	// location in Block.
	for i, l := range ploc.Block {
		if l.UID == loc.UID {
			// When location is the first element, it is necessary to check the dependency of its parent location.
			if i == 0 {
				if data, err := f.getDependencyLocationUIDS(data, ploc); err != nil {
					return DependencyTask{}, err
				} else {
					return data, nil
				}
			}
			// When location is not the first element, dependency location is the preceding element in the same array.
			return DependencyTask{
				Tasks:    f.findAllTasks(ploc.Block[i-1]),
				Strategy: succeedExecuteStrategy,
			}, nil
		}
	}

	// location in Rescue
	for i, l := range ploc.Rescue {
		if l.UID == loc.UID {
			// When location is the first element, dependency location is all task of sibling block array.
			if i == 0 {
				return DependencyTask{
					Tasks:    f.findAllTasks(ploc.Block[len(ploc.Block)-1]),
					Strategy: failedExecuteStrategy,
				}, nil
			}
			// When location is not the first element, dependency location is the preceding element in the same array
			return DependencyTask{
				Tasks:    f.findAllTasks(ploc.Rescue[i-1]),
				Strategy: succeedExecuteStrategy}, nil
		}
	}

	// If location in Always
	for i, l := range ploc.Always {
		if l.UID == loc.UID {
			// When location is the first element, dependency location is all task of sibling block array
			if i == 0 {
				return DependencyTask{
					Tasks:    f.findAllTasks(ploc.Block[len(ploc.Block)-1]),
					Strategy: alwaysExecuteStrategy,
				}, nil
			}
			// When location is not the first element, dependency location is the preceding element in the same array
			return DependencyTask{
				Tasks:    f.findAllTasks(ploc.Always[i-1]),
				Strategy: alwaysExecuteStrategy,
			}, nil

		}
	}

	return DependencyTask{}, fmt.Errorf("connot find location %s in parent %s", loc.UID, loc.PUID)
}

func (f DependencyTasks) findAllTasks(loc location) []string {
	if len(loc.Block) == 0 {
		return []string{loc.UID}
	}
	var result = make([]string, 0)
	for _, l := range loc.Block {
		result = append(result, f.findAllTasks(l)...)
	}
	for _, l := range loc.Rescue {
		result = append(result, f.findAllTasks(l)...)
	}
	for _, l := range loc.Always {
		result = append(result, f.findAllTasks(l)...)
	}

	return result
}

type MergeOption interface {
	mergeTo(data *value) error
}

// HostMerge merge variable to host
type HostMerge struct {
	// HostName of host
	HostNames []string
	// LocationVars to find block. Only merge the last level block.
	//LocationVars []string
	LocationUID string
	// Data to merge
	Data VariableData
}

func (h HostMerge) mergeTo(v *value) error {
	for _, name := range h.HostNames {
		hv := v.Hosts[name]
		if h.LocationUID == "" { // merge to host var
			hv.Vars = mergeVariables(h.Data, v.Hosts[name].Vars)
		} else { // merge to host runtime
			if hv.RuntimeVars == nil {
				hv.RuntimeVars = make(map[string]VariableData)
			}
			hv.RuntimeVars[h.LocationUID] = mergeVariables(v.Hosts[name].RuntimeVars[h.LocationUID], h.Data)
		}
		v.Hosts[name] = hv
	}
	return nil
}

type LocationType string

const (
	BlockLocation  LocationType = "block"
	AlwaysLocation LocationType = "always"
	RescueLocation LocationType = "rescue"
)

// LocationMerge merge variable to location
type LocationMerge struct {
	UID       string
	ParentUID string
	Type      LocationType
	Name      string
	Vars      VariableData
}

func (t LocationMerge) mergeTo(v *value) error {
	if t.ParentUID == "" {
		v.Location = append(v.Location, location{
			Name: t.Name,
			PUID: t.ParentUID,
			UID:  t.UID,
			Vars: t.Vars,
		})
		return nil
	}
	// find parent graph
	parentLocation := findLocation(v.Location, t.ParentUID)
	if parentLocation == nil {
		return fmt.Errorf("cannot find parent location %s", t.ParentUID)
	}

	switch t.Type {
	case BlockLocation:
		for _, loc := range parentLocation.Block {
			if loc.UID == t.UID {
				klog.Warningf("task graph %s already exist", t.UID)
				return nil
			}
		}
		parentLocation.Block = append(parentLocation.Block, location{
			Name: t.Name,
			PUID: t.ParentUID,
			UID:  t.UID,
			Vars: t.Vars,
		})
	case AlwaysLocation:
		for _, loc := range parentLocation.Always {
			if loc.UID == t.UID {
				klog.Warningf("task graph %s already exist", t.UID)
				return nil
			}
		}
		parentLocation.Always = append(parentLocation.Always, location{
			Name: t.Name,
			PUID: t.ParentUID,
			UID:  t.UID,
			Vars: t.Vars,
		})
	case RescueLocation:
		for _, loc := range parentLocation.Rescue {
			if loc.UID == t.UID {
				klog.Warningf("task graph %s already exist", t.UID)
				return nil
			}
		}
		parentLocation.Rescue = append(parentLocation.Rescue, location{
			Name: t.Name,
			PUID: t.ParentUID,
			UID:  t.UID,
			Vars: t.Vars,
		})
	default:
		return fmt.Errorf("unknown LocationType. only support block,always,rescue ")
	}

	return nil
}

// Cache is a cache for variable
var Cache = cgcache.NewStore(func(obj interface{}) (string, error) {
	v, ok := obj.(Variable)
	if !ok {
		return "", fmt.Errorf("cannot convert %v to variable", obj)
	}
	return v.Key(), nil
})

func GetVariable(o Options) (Variable, error) {
	vars, ok, err := Cache.GetByKey(string(o.Pipeline.UID))
	if err != nil {
		klog.V(5).ErrorS(err, "get variable error", "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	if ok {
		return vars.(Variable), nil
	}
	// add new variable to cache
	nv, err := New(o)
	if err != nil {
		klog.V(5).ErrorS(err, "create variable error", "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	if err := Cache.Add(nv); err != nil {
		klog.V(5).ErrorS(err, "add variable to store error", "pipeline", ctrlclient.ObjectKeyFromObject(&o.Pipeline))
		return nil, err
	}
	return nv, nil
}

func CleanVariable(p *kubekeyv1.Pipeline) {
	if _, ok, err := Cache.GetByKey(string(p.UID)); err == nil && ok {
		Cache.Delete(string(p.UID))
	}
}
