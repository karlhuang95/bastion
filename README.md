# 集群切换小工具

使用前提，网络没问题，如果使用key，请设置好自己本地对应的私钥

配置文件需要放到 ～/.s/s.yaml 这个文件下
创建配置文件
```shell
$ mkdir ~/.s
$ cat ~/.s/s.yaml

clusters:
  - name: k8s_cluster01
    command: ssh user@dotuimao-server
  - name: k8s_cluster02
    command: ssh user@chuangzuomao-server
  - name: k8s_cluster03
    command: ssh user@dtmapp-server
```


```shell
./s

                                                                               集 群 切 换 工 具
                                                     使 用  ↑   ↓   键 选 择 ， 回 车 执 行 ， Esc 退 出 ， 按  'q' 退 出
  k8s_cluster01
  k8s_cluster02
  k8s_cluster03
```
