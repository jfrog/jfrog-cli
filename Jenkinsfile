node {
    cleanWs()
    def architectures = [
            [pkg: 'jfrog-cli-linux-386', goos: 'linux', goarch: '386', fileExtention: ''],
            [pkg: 'jfrog-cli-linux-amd64', goos: 'linux', goarch: 'amd64', fileExtention: ''],
            [pkg: 'jfrog-cli-mac-386', goos: 'darwin', goarch: 'amd64', fileExtention: ''],
            [pkg: 'jfrog-cli-windows-amd64', goos: 'windows', goarch: 'amd64', fileExtention: '.exe']
    ]

    subject = 'jfrog'
    repo = 'jfrog-cli-go'
    sh 'rm -rf temp'
    sh 'mkdir temp'
    def goRoot = tool 'go-1.11'
    env.GOROOT="$goRoot"
    env.PATH+=":${goRoot}/bin"
    env.GO111MODULE="on"
    env.JFROG_CLI_OFFER_CONFIG="false"

    dir('temp') {
        cliWorkspace = pwd()
        sh "echo cliWorkspace=$cliWorkspace"
        stage('Clone') {
            sh 'git clone https://github.com/jfrog/jfrog-cli-go.git'
            dir("$repo") {
                if (BRANCH?.trim()) {
                    sh "git checkout $BRANCH"
                }
            }
        }

        stage('Go Install') {
            jfrogCliRepoDir = "${cliWorkspace}/${repo}/"
            jfrogCliDir = "${jfrogCliRepoDir}jfrog-cli/jfrog"
            sh "echo jfrogCliDir=$jfrogCliDir"

            stage('Go Install') {
                sh 'go version'
                dir("$jfrogCliRepoDir") {
                    sh './build.sh'
                }
            }

            sh 'mkdir builder'
            sh "mv $jfrogCliRepoDir/jfrog builder/"

            // Extract cli version
            sh 'builder/jfrog --version > version'
            version = readFile('version').trim().split(" ")[2]
            print "CLI version: $version"
        }

        if ("$EXECUTION_MODE".toString().equals("Publish packages")) {
            stage('Npm Publish') {
                print "publishing npm package"
                publishNpmPackage(jfrogCliRepoDir)
            }

            stage('Build and Publish Docker Image') {
                buildPublishDockerImage(version, jfrogCliRepoDir)
            }
        } else if ("$EXECUTION_MODE".toString().equals("Build CLI")) {
            print "publishing version: $version"
            publishCliVersion(architectures)
        }
    }
}

def publishCliVersion(architectures) {
    for (int i = 0; i < architectures.size(); i++) {
        def currentBuild = architectures[i]
        stage("Build ${currentBuild.pkg}") {
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
           builder/jfrog bt u $jfrogCliRepoDir/$fileName $subject/$repo/$pkg/$version /$version/$pkg/ --user=$USER_NAME --key=$KEY
        """
}

def buildAndUpload(goos, goarch, pkg, fileExtension) {
    def extension = fileExtension == null ? '' : fileExtension
    def fileName = "jfrog$fileExtension"
    dir("${jfrogCliRepoDir}") {
        env.GOOS="$goos"
        env.GOARCH="$goarch"
        sh "./build.sh $fileName"
    }

    uploadToBintray(pkg, fileName)
    sh "rm $jfrogCliRepoDir/$fileName"
}

def publishNpmPackage(jfrogCliRepoDir) {
    dir(jfrogCliRepoDir+'npm/') {
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
