#!/bin/bash -xe

apt-get update

# Install nginx for instance http health check
apt-get install -y nginx
### missing in AWS ###
apt-get install -y libcurl3

#### install mms-automation-agent and mms-monitoring-agent ####
curl -OL https://cloud.mongodb.com/download/agent/automation/mongodb-mms-automation-agent-manager_latest_amd64.ubuntu1604.deb
dpkg -i mongodb-mms-automation-agent-manager_latest_amd64.ubuntu1604.deb

#### configure the mmsGroupId and mmsApiKey ####
sed  -i.bak "s|mmsGroupId=.*|mmsGroupId=${mmsGroupId}|g" /etc/mongodb-mms/automation-agent.config
sed  -i.bak "s|mmsApiKey=.*|mmsApiKey=${mmsApiKey}|g" /etc/mongodb-mms/automation-agent.config

### disable Transparent Huge Pages (THP) in Ubuntu 16.04LTS ###
echo never > /sys/kernel/mm/transparent_hugepage/enabled
echo never > /sys/kernel/mm/transparent_hugepage/defrag
echo 'echo never > /sys/kernel/mm/transparent_hugepage/enabled' | sudo tee --append /etc/rc.local
echo 'echo never > /sys/kernel/mm/transparent_hugepage/defrag' | sudo tee --append /etc/rc.local

### create xfs for /data ###
mkfs.xfs -L mongodb /dev/nvme1n1
echo -e "LABEL=mongodb   /data   xfs   defaults,noatime,discard        0 0" >> /etc/fstab
mkdir /data
mount -a
chown mongodb:mongodb /data
systemctl restart mongodb-mms-automation-agent.service