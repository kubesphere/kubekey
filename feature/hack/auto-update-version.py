#!/usr/bin/env python3
# encoding: utf-8

# Copyright 2022 The KubeSphere Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import requests
import re
import json
from natsort import natsorted
import collections

GITHUB_BASE_URL = "https://api.github.com"
ORG = "kubernetes"
REPO = "kubernetes"
PER_PAGE = 15

ARCH_LIST = ["amd64", "arm64"]
K8S_COMPONENTS = ["kubeadm", "kubelet", "kubectl"]


def get_releases(org, repo, per_page=30):
    try:
        response = requests.get("{}/repos/{}/{}/releases?per_page={}".format(GITHUB_BASE_URL, org, repo, per_page))
    except:
        print("fetch {}/{} releases failed".format(org, repo))
    else:
        return response.json()


def get_new_kubernetes_version(current_version):
    new_versions = []

    kubernetes_release = get_releases(org=ORG, repo=REPO, per_page=PER_PAGE)

    for release in kubernetes_release:
        tag = release['tag_name']
        res = re.search("^v[0-9]+.[0-9]+.[0-9]+$", tag)
        if res and tag not in current_version['kubeadm']['amd64'].keys():
            new_versions.append(tag)

    return new_versions


def fetch_kubernetes_sha256(versions):
    new_sha256 = {}

    for version in versions:
        for binary in K8S_COMPONENTS:
            for arch in ARCH_LIST:
                response = requests.get(
                    "https://storage.googleapis.com/kubernetes-release/release/{}/bin/linux/{}/{}.sha256".format(
                        version, arch, binary))
                if response.status_code == 200:
                    new_sha256["{}-{}-{}".format(binary, arch, version)] = response.text

    return new_sha256


def version_sort(data):
    version_list = natsorted([*data])
    sorted_data = collections.OrderedDict()

    for v in version_list:
        sorted_data[v] = data[v]

    return sorted_data


def main():
    # get current support versions
    with open("version/components.json", "r") as f:
        data = json.load(f)

    # get new kubernetes versions
    new_versions = get_new_kubernetes_version(current_version=data)

    if len(new_versions) > 0:
        # fetch new kubernetes sha256
        new_sha256 = fetch_kubernetes_sha256(new_versions)

        if new_sha256:
            for k, v in new_sha256.items():
                info = k.split('-')
                data[info[0]][info[1]][info[2]] = v

        for binary in K8S_COMPONENTS:
            for arch in ARCH_LIST:
                data[binary][arch] = version_sort(data[binary][arch])

        print(new_versions)
        # update components.json
        with open("version/components.json", 'w') as f:
            json.dump(data, f, indent=4, ensure_ascii=False)

        # set new version to tmp file
        with open("version.tmp", 'w') as f:
            f.write("\n".join(new_versions))


if __name__ == '__main__':
    main()
