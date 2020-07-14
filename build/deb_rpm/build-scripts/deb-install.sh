wget -qO - https://releases.jfrog.io/artifactory/api/gpg/key/public | apt-key add -;
echo "deb https://releases.jfrog.io/artifactory/jfrog-debs xenial contrib" | sudo tee -a /etc/apt/sources.list;
apt-get update; 
sudo apt-get install jfrog-cli