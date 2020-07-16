node {
    cleanWs()
    def architectures = [
            [pkg: 'jfrog-cli-windows-amd64', goos: 'windows', goarch: 'amd64', fileExtention: '.exe'],
            [pkg: 'jfrog-cli-linux-386', goos: 'linux', goarch: '386', fileExtention: '', debianImage: 'i386/ubuntu:19.10', debianArch: 'i386'],
            [pkg: 'jfrog-cli-linux-amd64', goos: 'linux', goarch: 'amd64', fileExtention: '', debianImage: 'ubuntu:19.10', debianArch: 'x86_64', rpmImage: 'centos:8'],
            [pkg: 'jfrog-cli-linux-arm64', goos: 'linux', goarch: 'arm64', fileExtention: ''],
            [pkg: 'jfrog-cli-linux-arm', goos: 'linux', goarch: 'arm', fileExtention: ''],
            [pkg: 'jfrog-cli-mac-386', goos: 'darwin', goarch: 'amd64', fileExtention: '']
    ]

    subject = 'jfrog'
    repo = 'jfrog-cli'
    sh 'rm -rf temp'
    sh 'mkdir temp'
    def goRoot = tool 'go-1.14'
    env.GOROOT="$goRoot"
    env.PATH+=":${goRoot}/bin"
    env.GO111MODULE="on"
    env.JFROG_CLI_OFFER_CONFIG="false"

    dir('temp') {
        cliWorkspace = pwd()
        sh "echo cliWorkspace=$cliWorkspace"
        stage('Clone JFrog CLI sources') {
            sh 'git clone https://github.com/jfrog/jfrog-cli.git'
            dir("$repo") {
                if (BRANCH?.trim()) {
                    sh "git checkout $BRANCH"
                }
            }
        }

        stage('Build JFrog CLI') {
            jfrogCliRepoDir = "${cliWorkspace}/${repo}/"
            jfrogCliDir = "${jfrogCliRepoDir}jfrog-cli/jfrog"
            sh "echo jfrogCliDir=$jfrogCliDir"

            sh 'go version'
            dir("$jfrogCliRepoDir") {
                sh 'build/build.sh'
            }

            sh 'mkdir builder'
            sh "mv $jfrogCliRepoDir/jfrog builder/"

            // Extract CLI version
            version = sh(script: "builder/jfrog -v | tr -d 'jfrog version' | tr -d '\n'", returnStdout: true)
            print "CLI version: $version"
        }

        if ("$EXECUTION_MODE".toString().equals("Publish packages")) {
            buildRpmAndDeb(version, architectures)

            stage('Npm Publish') {
                publishNpmPackage(jfrogCliRepoDir)
            }

            stage('Build and Publish Docker Image') {
                buildPublishDockerImage(version, jfrogCliRepoDir)
            }
        } else if ("$EXECUTION_MODE".toString().equals("Build CLI")) {
            downloadToolsCert()
            print "Uploading version $version to Bintray"
            uploadCli(architectures)
        }
    }
}

def downloadToolsCert() {
    stage('Download tools cert') {
        // Download the certificate file and key file, used for signing the JFrog CLI binary.
        sh """#!/bin/bash
            builder/jfrog rt dl installation-files/certificates/jfrog/ --url https://entplus.jfrog.io/artifactory --flat --access-token=$DOWNLOAD_SIGNING_CERT_ACCESS_TOKEN
            """

        sh 'tar xvzf jfrogltd_signingcer_full.tar.gz'
    }
}

def buildRpmAndDeb(version, architectures) {
    boolean built = false
    for (int i = 0; i < architectures.size(); i++) {
        def currentBuild = architectures[i]
        if (currentBuild.debianImage) {
            stage("Build debian ${currentBuild.pkg}") {
                build(currentBuild.goos, currentBuild.goarch, currentBuild.pkg, 'jfrog')
                dir("$jfrogCliRepoDir") {
                    sh "build/deb_rpm/build-scripts/pack.sh -b jfrog -v $version -f deb --deb-arch $currentBuild.debianArch --deb-build-image $currentBuild.debianImage -t --deb-test-image $currentBuild.debianImage"
                    built = true
                }
            }
        }
        if (currentBuild.rpmImage) {
            stage("Build rpm ${currentBuild.pkg}") {
                build(currentBuild.goos, currentBuild.goarch, currentBuild.pkg, 'jfrog')
                dir("$jfrogCliRepoDir") {
                    sh "build/deb_rpm/build-scripts/pack.sh -b jfrog -v $version -f rpm --rpm-build-image $currentBuild.rpmImage -t --rpm-test-image $currentBuild.rpmImage"
                    built = true
                }
            }
        }
    }

    if (built) {
        stage("Deploy deb and rpm") {
            options = "--url https://releases.jfrog.io/artifactory --flat --access-token=$DEB_RPM_DEPLOY_ACCESS_TOKEN"
            sh """#!/bin/bash
                builder/jfrog rt u $jfrogCliRepoDir/build/deb_rpm/*.i386.deb jfrog-debs/pool/jfrog-cli/ --deb=xenial,bionic,eoan,focal/contrib/i386 $options
                builder/jfrog rt u $jfrogCliRepoDir/build/deb_rpm/*.x86_64.deb jfrog-debs/pool/jfrog-cli/ --deb=xenial,bionic,eoan,focal/contrib/amd64 $options
                builder/jfrog rt u $jfrogCliRepoDir/build/deb_rpm/*.rpm jfrog-rpms/jfrog-cli/ $options
                """
        }
    } 
}

def uploadCli(architectures) {
    for (int i = 0; i < architectures.size(); i++) {
        def currentBuild = architectures[i]
        stage("Build and upload ${currentBuild.pkg}") {
            buildAndUpload(currentBuild.goos, currentBuild.goarch, currentBuild.pkg, currentBuild.fileExtention)
        }
    }
}

def buildPublishDockerImage(version, jfrogCliRepoDir) {
    dir("$jfrogCliRepoDir") {
        docker.build("jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:$version")
        sh '#!/bin/sh -e\n' + 'echo $KEY | docker login --username=$USER_NAME --password-stdin jfrog-docker-reg2.bintray.io/jfrog'
        sh "docker push jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:$version"
        sh "docker tag jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:$version jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:latest"
        sh "docker push jfrog-docker-reg2.bintray.io/jfrog/jfrog-cli-go:latest"
    }
}

def uploadToBintray(pkg, fileName) {
    sh """#!/bin/bash
           builder/jfrog bt u $jfrogCliRepoDir/$fileName $subject/jfrog-cli-go/$pkg/$version /$version/$pkg/ --user=$USER_NAME --key=$KEY
        """
}

def build(goos, goarch, pkg, fileName) {
    dir("${jfrogCliRepoDir}") {
        env.GOOS="$goos"
        env.GOARCH="$goarch"
        sh "build/build.sh $fileName"

        if (goos == 'windows') {
            dir("${cliWorkspace}/certs-dir") {
                // Move the jfrog executable into the 'sign' directory, so that it is signed there.
                sh "mv $jfrogCliRepoDir/$fileName ${jfrogCliRepoDir}build/sign/${fileName}.unsigned"
                // Copy all the certificate files into the 'sign' directory.
                sh "cp * ${jfrogCliRepoDir}build/sign/"
                // Build and run the docker container, which signs the JFrog CLI binary.
                sh "docker build -t jfrog-cli-sign-tool ${jfrogCliRepoDir}build/sign/"
                def signCmd = "osslsigncode sign -certs workspace/JFrog_Ltd_.crt -key workspace/jfrogltd.key  -n JFrog_CLI -i https://www.jfrog.com/confluence/display/CLI/JFrog+CLI -in workspace/${fileName}.unsigned -out workspace/$fileName"
                sh "docker run -v ${jfrogCliRepoDir}build/sign/:/workspace --rm jfrog-cli-sign-tool $signCmd"
                // Move the JFrog CLI binary from the 'sign' directory, back to its original location.
                sh "mv ${jfrogCliRepoDir}build/sign/$fileName $jfrogCliRepoDir"
            }
        }
    }
}

def buildAndUpload(goos, goarch, pkg, fileExtension) {
    def extension = fileExtension == null ? '' : fileExtension
    def fileName = "jfrog$fileExtension"

    build(goos, goarch, pkg, fileName)
    uploadToBintray(pkg, fileName)
    sh "rm $jfrogCliRepoDir/$fileName"
}

def publishNpmPackage(jfrogCliRepoDir) {
    dir(jfrogCliRepoDir+'build/npm/') {
        sh '''#!/bin/bash
            echo "Downloading npm..."
            wget https://nodejs.org/dist/v8.11.1/node-v8.11.1-linux-x64.tar.xz
            tar -xvf node-v8.11.1-linux-x64.tar.xz
            export PATH=$PATH:$PWD/node-v8.11.1-linux-x64/bin/
            echo "//registry.npmjs.org/:_authToken=$NPM_AUTH_TOKEN" > .npmrc
            echo "registry=https://registry.npmjs.org" >> .npmrc
            ./node-v8.11.1-linux-x64/bin/npm publish
        '''
    }
}
