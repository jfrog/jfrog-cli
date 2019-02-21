validateNpmVersion();

var https = require('https');
var http = require('http');
var fs = require('fs');
var packageJson = require('./package.json');
var fileName = getFileName();
var filePath = "bin/" + fileName;
var version = packageJson.version;
var btPackage = "jfrog-cli-" + getArchitecture();

downloadCli();

function validateNpmVersion() {
    if (!isValidNpmVersion()) {
        throw new Error("JFrog CLI can be installed using npm version 5.0.0 or above.")
    }
}

function download(url) {
    // We detect proxy by looking at the environment variable
    if(process.env.https_proxy && process.env.https_proxy.length > 0) {
        url = process.env.https_proxy + starturl.replace('https://', '/https/');
    }

    https.get(url, function(res) {
        if(res.statusCode == 301 || res.statusCode == 302) {
            download(res.headers.location);
        } else if (res.statusCode == 200) {
            writeToFile(res)
        } else {
            console.log('Unexpected status code during JFrog CLI download')
        }
    }).on('error', function (err) {console.error(err);});

}

function downloadCli() {
    console.log("Downloading JFrog CLI " + version );
    var startUrl = 'https://api.bintray.com/content/jfrog/jfrog-cli-go/' + version + '/' + btPackage + '/' + fileName + '?bt_package=' + btPackage;
    download(startUrl);
}

function isValidNpmVersion() {
    var child_process = require('child_process');
    var npmVersionCmdOut = child_process.execSync("npm version -json");
    var npmVersion = JSON.parse(npmVersionCmdOut).npm;
    // Supported since version 5.0.0
    return parseInt(npmVersion.charAt(0)) > 4
}

function writeToFile(response) {
    var file = fs.createWriteStream(filePath);
    response.on('data', function (chunk) {
        file.write(chunk);
    }).on('end', function () {
        file.end();
        if (!process.platform.startsWith("win")) {
            fs.chmodSync(filePath, 0555)
        }
    }).on('error', function (err) {
        console.error(err);
    });
}

function getArchitecture() {
    var platform = process.platform;
    if (platform.startsWith("win")) {
        return "windows-amd64"
    }
    if (platform.includes("darwin")) {
        return "mac-386"
    }
    if (process.arch.includes("64")) {
        return "linux-amd64"
    }
    return "linux-386"
}

function getFileName() {
    var excecutable = "jfrog";
    if (process.platform.startsWith("win")) {
        excecutable += ".exe"
    }
    return excecutable
}
