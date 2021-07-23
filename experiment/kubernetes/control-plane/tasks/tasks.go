package tasks

import (
   "github.com/kubesphere/kubekey/apis/kubekey/v1alpha1"
   "github.com/kubesphere/kubekey/experiment/utils/pipline"
   "github.com/kubesphere/kubekey/pkg/util/manager"
   "text/template"

)

var (
   initClusterCmd = "kubeadm init -f kubeadm.conf"
   getKubeConfigCmd = "mkdir -p /root/.kube && mkdir -p $HOME/.kube"
   addNodeCmdTmpl = template.Must(template.New("").Parse(
      "kubeadm join {{ .ApiServer }} --token {{ .Token }} --discovery-token-ca-cert-hash {{ .Hash }}",
      ))

   mgr = manager.Manager{}
   InitCluster = pipline.Task{
      Name:        "Init Cluster",
      Hosts:       []v1alpha1.HostCfg{mgr.MasterNodes[0]},
      Action: &pipline.Command{
         Cmd: initClusterCmd,
      },
      Env:         nil,
      Result:      pipline.Result{},
      Parallel:    false,
      Condition:   false,
      IgnoreError: false,
   }

   varsFunc = func(result pipline.Result) []map[string]interface{} {
      return nil
   } (InitCluster.Result)

   conditionFunc = func(result pipline.Result) bool {
      return true
   }(InitCluster.Result)


   GetKubeConfig = pipline.Task{
      Name:        "Get KubeConfig",
      Hosts:       mgr.MasterNodes,
      Action:      &pipline.Command{Cmd: getKubeConfigCmd},
      Vars:  varsFunc,
      Condition: conditionFunc,
   }
)
