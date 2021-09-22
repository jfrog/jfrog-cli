#!/bin/bash

# This file is responsible for building rpm and deb package for jfrog-cli installer

# This will contain hold the list of supported architectures which can be built by default.
# Although by passing a different --rpm-build-image or --rpm-build-image, artifacts of different architectures can be built
SUPPORTED_DEFAULT_ARCH_LIST="x86_64"

JFROG_CLI_HOME="$(cd "$(dirname "${BASH_SOURCE[0]}")/../" && pwd)"
JFROG_CLI_PKG="$JFROG_CLI_HOME/pkg"
JFROG_CLI_PREFIX="jfrog-cli"

RPM_BUILDER_NAME="jfrog-cli-rpm-builder"
RPM_IMAGE_ROOT_DIR="/opt/jfrog-cli"
RPM_DEPS="rpm-build"

DEB_BUILDER_NAME="jfrog-cli-deb-builder_3"
DEB_IMAGE_ROOT_DIR="/opt/jfrog-cli"
DEB_DEPS="devscripts build-essential lintian debhelper"

usage () {
    echo "Usage: ${0}"
    cat << END_USAGE

-b | --binary                                     : [mandatory] jfrog cli binary to be packaged in rpm/deb
-v | --version                                    : [mandatory] version of rpm/deb to be built
-t | --test                                       : [optional] test the generated rpm and deb
--rpm-build-image                                 : [optional] docker image to be used to build rpm
--deb-build-image                                 : [optional] docker image to be used to build deb
--rpm-test-image                                  : [optional] docker image to be used to test the generated rpm
--deb-test-image                                  : [optional] docker image to be used to test the generated deb
-f | --flavour                                    : [optional] flavours to be built (default : "rpm deb")
-h | --help                                       : [optional] display usage

END_USAGE
    sleep .5
}

log(){
	echo "$1"
}

errorExit() {
    echo -e "\033[31mERROR: $1 \033[0m"; echo
    exit 1
}

checkDockerAccess() {
	log "Checking docker" "DEBUG"
	docker -v > /dev/null 2>&1 && docker ps > /dev/null 2>&1
	if [ $? -ne 0 ]; then
	    errorExit "Must run as user that can execute docker commands"
	fi
	log "Docker is avaiable" "DEBUG"
}

exitWithUsage(){
	log "ERROR : $1"
	usage
	exit 1
}

createDEBPackage(){
	local flavour="deb"

	# cleanup old files and containers
	rm -f ${JFROG_CLI_PKG}/${JFROG_CLI_PREFIX}*${VERSION_FORMATTED}*.${flavour}
	docker rm -f "${RPM_BUILDER_NAME}" 2>/dev/null

	log "Building ${JFROG_CLI_PREFIX} ${flavour} ${JFROG_CLI_VERSION} on ${DEB_BUILD_IMAGE} image"

    docker run --rm -v "${JFROG_CLI_HOME}/${flavour}":${DEB_IMAGE_ROOT_DIR}/src \
					-v "${JFROG_CLI_PKG}":${DEB_IMAGE_ROOT_DIR}/pkg \
					--name ${DEB_BUILDER_NAME} \
							${DEB_BUILD_IMAGE} bash -c "\
										\
										echo '' && echo '' && \
										apt-get update && \
										apt-get install -y fakeroot && \
										DEBIAN_FRONTEND=noninteractive apt-get install -y ${DEB_DEPS} \
										--no-install-recommends tzdata && \
										echo '' && echo '' && \
									\
									\
										cp -fr ${DEB_IMAGE_ROOT_DIR}/src ${DEB_IMAGE_ROOT_DIR}-build && \
										cd ${DEB_IMAGE_ROOT_DIR}-build && \
										sed -i -e 's|__VERSION__|'${JFROG_CLI_VERSION}'|g' ${DEB_IMAGE_ROOT_DIR}-build/debian/changelog && \
										\
										\
										debuild -us -uc && \
										\
										\
										mkdir -p ${DEB_IMAGE_ROOT_DIR}/pkg && \
										cp -f ${DEB_IMAGE_ROOT_DIR}-build/../${JFROG_CLI_PREFIX}*.${flavour} \
											  ${DEB_IMAGE_ROOT_DIR}/pkg/${JFROG_CLI_PREFIX}-${VERSION_FORMATTED}.${flavour} && \
									\
									\
										echo '' && echo '' && \
										echo '############ Build successful ###################' && \
										ls -ltr ${DEB_IMAGE_ROOT_DIR}/pkg/ | grep ${flavour} && \
										echo '#################################################' || \
										exit 1 \
									" || errorExit "------------- Build Failed ! ------------------"
}

createRPMPackage(){
	local flavour="rpm"

	# cleanup old files and containers
	rm -f ${JFROG_CLI_PKG}/${JFROG_CLI_PREFIX}*${VERSION_FORMATTED}*.${flavour}
	docker rm -f "${RPM_BUILDER_NAME}" 2>/dev/null

	log "Building ${JFROG_CLI_PREFIX} ${flavour} ${JFROG_CLI_VERSION} on ${RPM_BUILD_IMAGE} image"

    docker run --rm -v "${JFROG_CLI_HOME}/${flavour}":${RPM_IMAGE_ROOT_DIR}/src \
					-v "${JFROG_CLI_PKG}":${RPM_IMAGE_ROOT_DIR}/pkg \
					--name ${RPM_BUILDER_NAME} \
							${RPM_BUILD_IMAGE} bash -c "\
										echo '' && echo '' && \
										yum install -y ${RPM_DEPS} && \
										echo '' && echo '' && \
									\
									\
										rpmbuild -bb \
											--define='_tmppath ${RPM_IMAGE_ROOT_DIR}/tmp' \
											--define='_topdir ${RPM_IMAGE_ROOT_DIR}' \
											--define='_rpmdir ${RPM_IMAGE_ROOT_DIR}/pkg' \
											--define='buildroot ${RPM_IMAGE_ROOT_DIR}/BUILDROOT' \
											--define='_sourcedir ${RPM_IMAGE_ROOT_DIR}/src' \
											--define='cli_version '${JFROG_CLI_VERSION} \
											--define='cli_release '${JFROG_CLI_RELEASE_VERSION} \
											--define='filename_prefix '${JFROG_CLI_PREFIX} \
											--define='build_arch '${JFROG_CLI_RPM_ARCH} \
											--define='full_version '${VERSION_FORMATTED} \
											${RPM_IMAGE_ROOT_DIR}/src/jfrog-cli.spec && \
									\
									\
										echo '' && echo '' && \
										echo '############ Build successful ###################' && \
										ls -ltr ${RPM_IMAGE_ROOT_DIR}/pkg/ | grep ${flavour} && \
										echo '#################################################' || \
										exit 1 \
									" || errorExit "------------- Build Failed ! ------------------"

}

rpmSign()(
  local flavour=rpm
  local fileName="${JFROG_CLI_PREFIX}-${VERSION_FORMATTED}.${flavour}"
  local filePath="${JFROG_CLI_PKG}"/${fileName}
  local filePathInImage="/opt/${fileName}"
  local keYID="${RPM_SIGN_KEY_ID}"
  local passphrase="${RPM_SIGN_PASSPHRASE}"
  local gpgFileInImage="/opt/${RPM_SIGN_KEY_NAME}"
  local gpgFileInHost="${JFROG_CLI_PKG}/${RPM_SIGN_KEY_NAME}"
  local rpmSignScript="rpm-sign.sh"


	if [[ -f "${filePath}" && -f "${gpgFileInHost}" ]]; then
		log ""; log "";
		log "Initiating rpm sign on ${filePath}..."
		docker run --rm -it --name cli-rpm-sign -v "${filePath}":${filePathInImage} \
            -v "${gpgFileInHost}":"${gpgFileInImage}" \
            -v "${JFROG_CLI_HOME}/build-scripts":${RPM_IMAGE_ROOT_DIR}/src \
            ${RPM_SIGN_IMAGE} \
                bash -c "yum install -y expect rpm-sign pinentry && \
                        ${RPM_IMAGE_ROOT_DIR}/src/${rpmSignScript} \"${gpgFileInImage}\" \"${keYID}\" \"${passphrase}\" \"${filePathInImage}\" \
                        && exit 0 || exit 1" \
            || { echo "ERROR: ############### RPM Sign Failed! ###################"; exit 1; }
		log "############## RPM is signed! ##################"
		log ""; log "";
	else
		echo "ERROR: Could not find ${filePath} or ${gpgFileInHost}"
		exit 1
	fi
)

runTests()(
	local flavour=$1

	[ ! -z "${flavour}" ] || { echo "Flavour is mandatory to run tests"; exit 1; }

	local fileName="${JFROG_CLI_PREFIX}-${VERSION_FORMATTED}.${flavour}"
	local filePath="${JFROG_CLI_PKG}"/${fileName}
	local testImage=""
	local installCommand=""
	local filePathInImage="/opt/${fileName}"
	local signatureTestCommand=true

	if [[ "${flavour}" == "rpm" ]]; then
		testImage="${RPM_TEST_IMAGE}"
		installCommand="rpm -ivh ${filePathInImage}"
		signatureTestCommand="rpm -qi ${filePathInImage} | grep 'Signature   : DSA'"
	else
		testImage="${DEB_TEST_IMAGE}"
		installCommand="dpkg -i ${filePathInImage}"
	fi

	if [ -f "${filePath}" ]; then
		log ""; log "";
		log "Testing ${filePath} on ${testImage}..."
		docker run --rm -it --name cli-test -v "${filePath}":${filePathInImage} ${testImage} \
			bash -c "${installCommand}       && jfrog -version | grep ${JFROG_CLI_VERSION} && \
			         ${signatureTestCommand} && exit 0 || exit 1" \
				|| { echo "ERROR: ############### Test failed! ###################"; exit 1; }
		log "############## Test passed ##################"
		log ""; log "";
	else
		echo "ERROR: Could not find ${filePath} to run test"
		exit 1
	fi
)

getArch(){
	local image=$1

	[ ! -z "$image" ] || return 0;

	docker run --rm ${image} bash -c "uname -m 2>/dev/null" 2>/dev/null
}

createPackage(){
	local flavour=$1

	[ ! -z "${flavour}" ] || errorExit "Flavour is not passed to createPackage method"

	cp -f "${JFROG_CLI_BINARY}" "${JFROG_CLI_HOME}"/${flavour}/jfrog \
		|| errorExit "Failed to copy ${JFROG_CLI_BINARY} to ${JFROG_CLI_HOME}/${flavour}/jfrog"


	case "$flavour" in
		rpm)
			[ ! -z "${JFROG_CLI_RPM_ARCH}" ] || JFROG_CLI_RPM_ARCH=$(getArch "${RPM_BUILD_IMAGE}")
			VERSION_FORMATTED="${JFROG_CLI_VERSION}.${JFROG_CLI_RPM_ARCH}"
			createRPMPackage
		;;
		deb)
			[ ! -z "${JFROG_CLI_DEB_ARCH}" ] || JFROG_CLI_DEB_ARCH=$(getArch "${DEB_BUILD_IMAGE}")
			VERSION_FORMATTED="${JFROG_CLI_VERSION}.${JFROG_CLI_DEB_ARCH}"
			createDEBPackage
		;;
		*)
			errorExit "Invalid flavour passed $flavour"
		;;
	esac
}

setBuildImage(){
	local arch="$1"

	[ ! -z "${arch}" ] || errorExit "Architecture is not passed to setBuildImage method"

	case "$1" in
		x86_64)
			RPM_BUILD_IMAGE="centos:7"
			DEB_BUILD_IMAGE="ubuntu:16.04"
		;;
		*)
			errorExit "Provided architecture is not supported : $arch. Supported list [ ${SUPPORTED_DEFAULT_ARCH_LIST} ]"
		;;
	esac
}

main(){
    while [[ $# > 0 ]]; do
        case "$1" in
            -f | --flavour)
                flavours="$2"
                shift 2
            ;;
            -b | --binary)
                JFROG_CLI_BINARY="$2"
                shift 2
            ;;
            -v | --version)
                JFROG_CLI_VERSION="$2"
                shift 2
            ;;
            --arch)
				setBuildImage "$2"
                shift 2
            ;;
            --rpm-arch)
                JFROG_CLI_RPM_ARCH="$2"
                shift 2
            ;;
            --deb-arch)
                JFROG_CLI_DEB_ARCH="$2"
                shift 2
            ;;
            --rpm-build-image)
                RPM_BUILD_IMAGE="$2"
                shift 2
            ;;
            --deb-build-image)
                DEB_BUILD_IMAGE="$2"
                shift 2
            ;;
            --rpm-test-image)
                RPM_TEST_IMAGE="$2"
                shift 2
            ;;
            --deb-test-image)
                DEB_TEST_IMAGE="$2"
                shift 2
            ;;
            -t | --test)
                JFROG_CLI_RUN_TEST="true"
                shift 1
            ;;
            *)
                usage
                exit 1
            ;;
        esac
    done

	: ${flavours:="rpm deb"}
	: ${JFROG_CLI_RUN_TEST:="false"}
	: ${RPM_BUILD_IMAGE:="centos:8"}
	: ${RPM_SIGN_IMAGE:="centos:7"}
	: ${DEB_BUILD_IMAGE:="ubuntu:16.04"}
	: ${DEB_TEST_IMAGE:="${DEB_BUILD_IMAGE}"}
	: ${RPM_TEST_IMAGE:="${RPM_BUILD_IMAGE}"}
	: ${JFROG_CLI_RELEASE_VERSION:="1"}
	: ${RPM_SIGN_PASSPHRASE:=""}
	: ${RPM_SIGN_KEY_ID:="JFrog Inc."}
	: ${RPM_SIGN_KEY_NAME:="RPM-GPG-KEY-jfrog-cli"}


	[ ! -z "${JFROG_CLI_BINARY}" ]  || exitWithUsage "jfrog cli binary is not passed"
	[ -f   "$JFROG_CLI_BINARY" ]    || exitWithUsage "jfrog cli binary is not available at $JFROG_CLI_BINARY"
	[ ! -z "${JFROG_CLI_VERSION}" ] || exitWithUsage "version is not passed, pass the version to be built"

  if [[ "$flavours" == *"rpm"* ]] && [[ -z "${RPM_SIGN_PASSPHRASE}" || "${RPM_SIGN_PASSPHRASE}" == "" ]]; then
    echo "ERROR: RPM_SIGN_PASSPHRASE environment variable is not set"
    exit 1
  fi

	log "Flavours being built are: $flavours"
	log "Version being built is $JFROG_CLI_VERSION"

	checkDockerAccess

	for flavour in $flavours; do
    createPackage "$flavour"
    [[ "${flavour}" == "rpm" ]]             && rpmSign               || true
    [[ "${JFROG_CLI_RUN_TEST}" == "true" ]] && runTests "${flavour}" || true
	done

	log "...and Done!"
}

main "$@"
exit 0
