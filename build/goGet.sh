GO111MODULE=on go get github.com/jfrog/jfrog-cli;
if [ -z "$GOPATH" ]
then
    binPath="$HOME/go/bin";
else
    binPath="$GOPATH/bin";
fi;
mv "$binPath/jfrog-cli" "$binPath/jfrog";
echo "$($binPath/jfrog -v) is installed at $binPath";
