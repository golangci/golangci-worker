#!/bin/bash
set -x
set -e

mkdir -p .ssh
ssh-keyscan -H github.com >>.ssh/known_hosts
ssh-keyscan -H $(cat $ENV_DIR/REMOTE_SHELL_HOST) >>.ssh/known_hosts
echo added github.com and remote shell to known hosts

echo "i am in $(pwd)"
echo "ENV_DIR is $ENV_DIR"
ls -l $ENV_DIR

cp $ENV_DIR/GIT_DEPLOY_KEY .ssh/id_rsa
chmod 0600 .ssh/id_rsa

cp $ENV_DIR/REMOTE_SHELL_KEY .ssh/remote_shell_key
chmod 0600 .ssh/remote_shell_key

echo added ssh keys

echo installing linters binaries
mkdir -p bin
wget https://s3-us-west-2.amazonaws.com/golangci-linters/v1/bin.tar.gz -O - | tar -C bin -xzvf -

mkdir -p .profile.d
echo 'PATH=$PATH:/app/src/github.com/golangci/golangci-worker/bin' > .profile.d/golangci.sh
echo successfuly installed linters binaries
