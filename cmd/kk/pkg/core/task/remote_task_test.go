/*
 Copyright 2021 The KubeSphere Authors.

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

package task

import (
	"testing"

	"github.com/kubesphere/kubekey/v3/cmd/kk/pkg/core/connector"
)

func TestTask_calculateConcurrency(t1 *testing.T) {
	type fields struct {
		Hosts       []connector.BaseHost
		Concurrency float64
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name: "test1",
			fields: fields{
				Concurrency: 0.5,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
				},
			},
			want: 2,
		},
		{
			name: "test2",
			fields: fields{
				Concurrency: 0.5,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 2,
		},
		{
			name: "test3",
			fields: fields{
				Concurrency: 0.4,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
				},
			},
			want: 1,
		},
		{
			name: "test4",
			fields: fields{
				Concurrency: 0.4,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 2,
		},
		{
			name: "test5",
			fields: fields{
				Concurrency: 0.1,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 1,
		},
		{
			name: "test6",
			fields: fields{
				Concurrency: 0.222222222222222,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 1,
		},
		{
			name: "test7",
			fields: fields{
				Concurrency: 1,
				Hosts: []connector.BaseHost{
					{Name: "node1"},
					{Name: "node2"},
					{Name: "node3"},
					{Name: "node4"},
				},
			},
			want: 4,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			var hosts []connector.Host
			for _, v := range tt.fields.Hosts {
				hosts = append(hosts, &v)
			}

			t := &RemoteTask{
				Concurrency: tt.fields.Concurrency,
				Hosts:       hosts,
			}
			if got := t.calculateConcurrency(); got != tt.want {
				t1.Errorf("calculateConcurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}
