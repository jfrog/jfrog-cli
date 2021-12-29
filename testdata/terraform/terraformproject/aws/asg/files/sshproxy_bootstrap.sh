#!/bin/bash -xe

# Enable ip forwarding and nat
sysctl -w net.ipv4.ip_forward=1

# Make forwarding persistent.
sed -i= 's/^[# ]*net.ipv4.ip_forward=[[:digit:]]/net.ipv4.ip_forward=1/g' /etc/sysctl.conf

iptables -t nat -A POSTROUTING -o ens5 -j MASQUERADE

apt-get update

# Install nginx for instance http health check
apt-get install -y nginx