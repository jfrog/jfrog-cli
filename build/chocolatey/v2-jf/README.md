# Building the Chocolatey package

Since the build is run in Jenkins in a Linux environment, the Chocolatey package
is built using the unofficial Docker image [linuturk/mono-choco][mono-choco]. The
official [Dockerfile][choco-dockerfile] was [contributed][choco-dockerfile-pr] by the
maintainer of the unofficial image and is pending a [PR][choco-image-pr] to become
the official Docker image.

To generate a package locally, either first compile or download the Windows executable
and place it in _build/chocolatey/tools_ directory.

With Docker:

```bash
cd build/chocolatey
docker run --rm -it -v $(pwd):/work -w /work linuturk/mono-choco pack version=<version>
```

With `choco` on Windows

```powershell
cd build\chocolatey
choco pack version=<version>
```

This will create the file _build/chocolatey/jfrog-cli.\<version\>.nupkg which can be
installed with Chcolatey

```powershell
choco install jfrog-cli.<version>.nupkg
```

See Chocolatey's official documenattion [here](https://chocolatey.org/docs/create-packages)

[choco-dockerfile-pr]: https://github.com/chocolatey/choco/pull/1153
[choco-dockerfile]: https://github.com/chocolatey/choco/tree/master/docker
[choco-image-pr]: https://github.com/chocolatey/choco/issues/1718
[mono-choco]: https://github.com/Linuturk/mono-choco/
