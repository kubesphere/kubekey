DOCKER_PLATFORM?=linux/arm64
DOCKER_BUILD_IMAGES?=false
IMAGENAMESPACE?=hub.kubesphere.com.cn/harbor
VERSION?=v2.10.2
BASEIMAGENAMESPACE?=hub.kubesphere.com.cn/harbor
SAVEIMAGENAMESPACE?=goharbor
BASEIMAGETAG?=v2.10.2
NPM_REGISTRY?=https://registry.npmmirror.com


.PHONY: build
build:
	if [ ! -d '_source' ];then \
	  mkdir _source; \
   	  git clone -b $(VERSION) https://github.com/goharbor/harbor.git _source/; \
   	  cd _source/ && git apply --ignore-space-change ../build_$(VERSION).patch; \
   	fi
	# 离线包自包含：docker save 进 tar 的镜像命名空间是 SAVEIMAGENAMESPACE(=goharbor, 即 docker.io/goharbor)，
	# 而镜像从 IMAGENAMESPACE(=hub.kubesphere.com.cn/harbor) 拉取后被重新打 tag 成 goharbor。
	# prepare 容器镜像也必须对齐 SAVEIMAGENAMESPACE，否则离线部署时 prepare 会去拉 hub 而本地只有 docker.io/goharbor。
	# 该对齐在 build_v2.10.2.patch 的 update_prepare_version 中完成（改用 $(SAVEIMAGENAMESPACE)），无需在此处再 sed。
	cd _source && make package_offline SAVEIMAGENAMESPACE=$(SAVEIMAGENAMESPACE) VERSIONTAG=$(VERSION) PKGVERSIONTAG=$(VERSION) BASEIMAGENAMESPACE=$(BASEIMAGENAMESPACE) IMAGENAMESPACE=$(IMAGENAMESPACE) BASEIMAGETAG=$(BASEIMAGETAG) DOCKER_PLATFORM=$(DOCKER_PLATFORM) DOCKER_BUILD_IMAGES=$(DOCKER_BUILD_IMAGES) TRIVYFLAG=true NPM_REGISTRY=$(NPM_REGISTRY)
	mv _source/harbor-offline-installer*.tgz . && rm -rf _source/
