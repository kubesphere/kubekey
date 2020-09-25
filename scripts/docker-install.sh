#!/bin/sh
set -e

docker_version=19.03.8
#apt_yum_url="https://download.docker.com"
apt_yum_url="https://mirrors.aliyun.com/docker-ce"

rhel_repos="
rhel-7-server-extras-rpms
rhui-REGION-rhel-server-extras
rhui-rhel-7-server-rhui-extras-rpms
rhui-rhel-7-for-arm-64-extras-rhui-rpms
"

mirror=''
while [ $# -gt 0 ]; do
        case "$1" in
                --mirror)
                        mirror="$2"
                        shift
                        ;;
                *)
                        echo "Illegal option $1"
                        ;;
        esac
        shift $(( $# > 0 ? 1 : 0 ))
done

case "$mirror" in
        AzureChinaCloud)
                apt_yum_url="https://mirror.azure.cn/docker-ce"
                ;;
        Aliyun)
                apt_yum_url="https://mirrors.aliyun.com/docker-ce"
                ;;
esac

command_exists() {
    command -v "$@" > /dev/null 2>&1
}

config_daemon(){
    sudo mkdir -p /etc/docker
    sudo cat > /etc/docker/daemon.json <<EOF
{
  "log-opts": {
    "max-size": "5m",
    "max-file":"3"
  },
  "exec-opts": ["native.cgroupdriver=systemd"],
}
EOF
    sudo systemctl daemon-reload
    sudo systemctl restart docker
}

echo_docker_as_nonroot() {
    if command_exists docker && [ -e /var/run/docker.sock ]; then
        (
            set -x
            $sh_c 'docker version'
        ) || true
    fi
    config_daemon

    your_user=your-user
    [ "$user" != 'root' ] && your_user="$user"
    # intentionally mixed spaces and tabs here -- tabs are stripped by "<<-EOF", spaces are kept in the output
    cat <<EOF
    If you would like to use Docker as a non-root user, you should now consider
    adding your user to the "docker" group with something like:
      sudo usermod -aG docker $your_user
    Remember that you will have to log out and back in for this to take effect!
    WARNING: Adding a user to the "docker" group will grant the ability to run
             containers which can be used to obtain root privileges on the
             docker host.
             Refer to https://docs.docker.com/engine/security/security/#docker-daemon-attack-surface
             for more information.
EOF
}
# Check if this is a forked Linux distro
check_forked() {

    # Check for lsb_release command existence, it usually exists in forked distros
    if command_exists lsb_release; then
        # Check if the `-u` option is supported
        set +e
        lsb_release -a -u > /dev/null 2>&1
        lsb_release_exit_code=$?
        set -e

        # Check if the command has exited successfully, it means we're in a forked distro
        if [ "$lsb_release_exit_code" = "0" ]; then
            # Print info about current distro
            cat <<EOF
            You're using '$lsb_dist' version '$dist_version'.
EOF

            # Get the upstream release info
            lsb_dist=$(lsb_release -a -u 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'id' | cut -d ':' -f 2 | tr -d '[[:space:]]')
            dist_version=$(lsb_release -a -u 2>&1 | tr '[:upper:]' '[:lower:]' | grep -E 'codename' | cut -d ':' -f 2 | tr -d '[[:space:]]')

            # Print info about upstream distro
            cat <<EOF
            Upstream release is '$lsb_dist' version '$dist_version'.
EOF
        else
            if [ -r /etc/debian_version ] && [ "$lsb_dist" != "ubuntu" ] && [ "$lsb_dist" != "raspbian" ]; then
                # We're Debian and don't even know it!
                lsb_dist=debian
                dist_version="$(cat /etc/debian_version | sed 's/\/.*//' | sed 's/\..*//')"
                case "$dist_version" in
                    10)
                        dist_version="buster"
                    ;;
                    9)
                        dist_version="stretch"
                    ;;
                    8|'Kali Linux 2')
                        dist_version="jessie"
                    ;;
                    7)
                        dist_version="wheezy"
                    ;;
                esac
            fi
        fi
    fi
}



do_install(){
    architecture=$(uname -m)
    case $architecture in
        # officially supported
        amd64|aarch64|arm64|x86_64)
            ;;
        # unofficially supported with available repositories
        armv6l|armv7l)
            ;;
        # unofficially supported without available repositories
        ppc64le|s390x)
            cat 1>&2 <<EOF
            Error: This install script does not support $architecture, because no
            $architecture package exists in Docker's repositories.
            Other install options include checking your distribution's package repository
            for a version of Docker, or building Docker from source.
EOF
            exit 1
            ;;
        # not supported
        *)
            cat >&2 <<EOF
            Error: $architecture is not a recognized platform.
EOF
            exit 1
            ;;
    esac
    if command_exists docker && [ -e /var/run/docker.sock ]; then
        version="$(docker -v | cut -d ' ' -f3 | cut -d ',' -f1)"
        echo $version
        
        cat >&2 <<EOF
        Warning: the "docker" command appears to already exist on this system.
        If you already have Docker installed, this script can cause trouble, which is
        why we're displaying this warning and exit the
        installation.
EOF
        ( set -x; sleep 20 )
  
    fi

    user="$(id -un 2>/dev/null || true)"

    sh_c='sh -c'
    if [ "$user" != 'root' ]; then
        if command_exists sudo; then
            sh_c='sudo -E sh -c'
        elif command_exists su; then
            sh_c='su -c'
        else
            cat >&2 <<EOF
            Error: this installer needs the ability to run commands as root.
            We are unable to find either "sudo" or "su" available to make this happen.
EOF
            exit 1
        fi
    fi

    curl=''
    if command_exists curl; then
        curl='curl -sSL'
    elif command_exists wget; then
        curl='wget -qO-'
    elif command_exists busybox && busybox --list-modules | grep -q wget; then
        curl='busybox wget -qO-'
    fi

    # check to see which repo they are trying to install from
    if [ -z "$repo" ]; then
        repo='main'
    fi

    # perform some very rudimentary platform detection
    lsb_dist=''
    dist_version=''
    if command_exists lsb_release; then
        lsb_dist="$(lsb_release -si)"
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/lsb-release ]; then
        sb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/debian_version ]; then
        lsb_dist='debian'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/fedora-release ]; then
        lsb_dist='fedora'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/oracle-release ]; then
        lsb_dist='oracleserver'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/centos-release ]; then
        lsb_dist='centos'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/redhat-release ]; then
        lsb_dist='redhat'
    fi
    if [ -z "$lsb_dist" ] && [ -r /etc/os-release ]; then
        lsb_dist="$(. /etc/os-release && echo "$ID")"
    fi

    lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"

    # Special case redhatenterpriseserver
    if [ "${lsb_dist}" = "redhatenterpriseserver" ]; then
            # Set it to redhat, it will be changed to centos below anyways
        lsb_dist='redhat'
    fi

    case "$lsb_dist" in
        ubuntu)
            if command_exists lsb_release; then
                dist_version="$(lsb_release --codename | cut -f2)"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/lsb-release ]; then
                dist_version="$(. /etc/lsb-release && echo "$DISTRIB_CODENAME")"
            fi
            ;;

        debian|raspbian)
            dist_version="$(cat /etc/debian_version | sed 's/\/.*//' | sed 's/\..*//')"
            case "$dist_version" in
                9)
                    dist_version="stretch"
                    ;;
                8)
                    dist_version="jessie"
                    ;;
                7)
                    dist_version="wheezy"
                    ;;
            esac
            ;;

        oracleserver)
                # need to switch lsb_dist to match yum repo URL
            lsb_dist="oraclelinux"
            dist_version="$(rpm -q --whatprovides redhat-release --queryformat "%{VERSION}\n" | sed 's/\/.*//' | sed 's/\..*//' | sed 's/Server*//')"
            ;;

        fedora|centos|redhat)
            dist_version="$(rpm -q --whatprovides ${lsb_dist}-release --queryformat "%{VERSION}\n" | sed 's/\/.*//' | sed 's/\..*//' | sed 's/Server*//' | sort | tail -1)"
            ;;

        *)
            if command_exists lsb_release; then
                dist_version="$(lsb_release --codename | cut -f2)"
            fi
            if [ -z "$dist_version" ] && [ -r /etc/os-release ]; then
                dist_version="$(. /etc/os-release && echo "$VERSION_ID")"
            fi
            ;;
    esac

        # Check if this is a forked Linux distro
    check_forked

    # Run setup for each distro accordingly
    case "$lsb_dist" in
        ubuntu|debian)
            pre_reqs="apt-transport-https ca-certificates curl"
            if [ "$lsb_dist" = "debian" ] && [ "$dist_version" = "wheezy" ]; then
                pre_reqs="$pre_reqs python-software-properties"
                backports="deb http://ftp.debian.org/debian wheezy-backports main"
                if ! grep -Fxq "$backports" /etc/apt/sources.list; then
                    (set -x; $sh_c "echo \"$backports\" >> /etc/apt/sources.list")
                fi
            else
                pre_reqs="$pre_reqs software-properties-common"
            fi
            if ! command -v gpg > /dev/null; then
                pre_reqs="$pre_reqs gnupg"
            fi
            apt_repo="deb [arch=$(dpkg --print-architecture)] ${apt_yum_url}/linux/$lsb_dist $dist_version stable"
            (
                set -x
                $sh_c 'apt-get update'
                $sh_c "apt-get install -y -q $pre_reqs"
                curl -fsSl "${apt_yum_url}/linux/$lsb_dist/gpg" | $sh_c 'apt-key add -'
                $sh_c "add-apt-repository \"$apt_repo\""
                if [ "$lsb_dist" = "debian" ] && [ "$dist_version" = "wheezy" ]; then
                        $sh_c 'sed -i "/deb-src.*download\.docker/d" /etc/apt/sources.list'
                fi
                $sh_c 'apt-get update'
                $sh_c "apt-get install -y -q docker-ce=$(apt-cache madison docker-ce | grep ${docker_version} | head -n 1 | cut -d ' ' -f 4)"
            )
            config_daemon
            exit 0
            ;;
        centos|fedora|redhat|oraclelinux)
            yum_repo="${apt_yum_url}/linux/centos/docker-ce.repo"
            if [ "$lsb_dist" = "fedora" ]; then
                if [ "$dist_version" -lt "24" ]; then
                    echo "Error: Only Fedora >=24 are supported by $url"
                    exit 1
                fi
                pkg_manager="dnf"
                config_manager="dnf config-manager"
                enable_channel_flag="--set-enabled"
                pre_reqs="dnf-plugins-core"
            else
                pkg_manager="yum"
                config_manager="yum-config-manager"
                enable_channel_flag="--enable"
                pre_reqs="yum-utils"
            fi
            (
                set -x
                if [ "$lsb_dist" = "redhat" ]; then
                    for rhel_repo in $rhel_repos ; do
                        $sh_c "$config_manager $enable_channel_flag $rhel_repo"
                    done
                    fi
                    $sh_c "$pkg_manager install -y -q $pre_reqs"
                    $sh_c "$config_manager --add-repo $yum_repo"
                    $sh_c "$pkg_manager makecache fast"
                    $sh_c "$pkg_manager install -y -q docker-ce-${docker_version}"
                    if [ -d '/run/systemd/system' ]; then
                        $sh_c 'service docker start'
                    else
                        $sh_c 'systemctl start docker'
                    fi
            )
            config_daemon
            exit 0
            ;;      
    esac

    # intentionally mixed spaces and tabs here -- tabs are stripped by "<<-'EOF'", spaces are kept in the output
    cat >&2 <<EOF
    Either your platform is not easily detectable or is not supported by this
    installer script.
    Please visit the following URL for more detailed installation instructions:
    https://docs.docker.com/engine/installation/
EOF
    
    exit 1
}

do_install




