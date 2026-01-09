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
	cd _source && make package_offline SAVEIMAGENAMESPACE=$(SAVEIMAGENAMESPACE) VERSIONTAG=$(VERSION) PKGVERSIONTAG=$(VERSION) BASEIMAGENAMESPACE=$(BASEIMAGENAMESPACE) IMAGENAMESPACE=$(IMAGENAMESPACE) BASEIMAGETAG=$(BASEIMAGETAG) DOCKER_PLATFORM=$(DOCKER_PLATFORM) DOCKER_BUILD_IMAGES=$(DOCKER_BUILD_IMAGES) TRIVYFLAG=true NPM_REGISTRY=$(NPM_REGISTRY)
	mv _source/harbor-offline-installer*.tgz . && rm -rf _source/
