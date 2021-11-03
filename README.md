基于k8s的算法平台一般对服务器性能都有很高的要求。我们在k8s上进行模型训练时发现，当出现个别服务器带宽挤兑时，会严重影响模型的整体训练速度。
本项目通过device-plugin上报当前节点的流量大小，通过scheduler中调度算法可以讲训练的worker节点调度到带宽充足的节点，避免发生个别节点
网络流量过大导致训练时间增加的问题。
启动方式：```go run ./cmd/server/app.go -i <eht0>```

network-device-plugin建议使用ds方式部署到集群中
```
      - env:
        - name: OS_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
```
network-device-plugin通过上面的环境变量可以获得当前使用的网卡，无需通过传参的方式传入网卡信息。

