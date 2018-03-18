# OpenShift Gochat Server
## Introduction
This is an incredibly light-weight chat room written in Golang.  There is no UI and it only exposes a websocket.

## Requirements
This requires the [golang-s2i builder image](https://github.com/kevensen/golang-s2i).  Build and deploy this in your project space first.

## Build and Deployment
The build and deployment relies on the [golang template](https://github.com/kevensen/golang-s2i/templates/golang.yml).  This can be accomplished with:
```terminal
oc process -f https://raw.githubusercontent.com/kevensen/golang-s2i/master/templates/golang.yml -p APP_ARGS='-logtostderr -host :8080' -p APP_SOURCE_REPOSITORY_URL=https://github.com/kevensen/openshift-gochat-server -p APPLICATION_NAME=gochat-server | oc create -f -
```

## Deployment Notes
It may be desirable to keep all chat traffic inside the OpenShift cluster and not expose the *gochat-server* route.  If this is the case, remove the route:
```terminal
oc delete route route gochat-server
```

By default, OpenShift install with the **redhat/openshift-ovs-subnet** network plugin.  This provides a flat network space and allows traffic between projects.  In this case, it will allow openshift-gochat-clients in various projects to communicate with the gochat-server in a different project.  

However, if your OpenShift cluister was deployed with the **redhat/openshift-ovs-multitenant** network plugin, network traffic is isolated between projects.  Fortunately OpenShift provides a mechanism to work around this issue; make the gochat-server project network global:
```terminal
oc adm pod-network make-projects-global gochat-server
```

## Final Notes
If you want to keep the chat traffic internal to the cluster, you'll need to note the FQDN of the gochat-server service.  If your gochat-server project is named **gochat-server**, the FQDN is **gochat-server.gochat-server.svc.cluster.local:8080**.