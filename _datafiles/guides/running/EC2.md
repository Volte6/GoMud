### Setting Up an EC2 Instance and Hosting GoMud

If you plan to host GoMUD on a remote server using AWS, EC2 is a solid option. Here are some tips and tricks to get you started.

Note: Configuration of the separate datafiles repository is optional, but be aware that different GoMud configuration options might be necessary

Before embarking on this journey, please familiarize yourself with the [general running guide](https://github.com/Volte6/GoMud/blob/master/_datafiles/guides/running/README.md).


#### Step 1: Create EC2 Instance

Set up an ec2 instance on your AWS account. Here are some suggested options:
* Use amazon linux
* t2.micro is acceptable, but consider t2.small
* Create a keypair. You will need this to ssh into the server.
* Create a security group and ensure that telnet, http, https, and ssh incoming traffic is enabled. Consider limiting the IP range, or adding a separate ACL.
* At time of writing the default storage volume is 8GiB. Consider if your deployment needs an additional volume.

Note: Registering a domain name and assigning a hostname is out of scope

Links:
* [AWS Documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EC2_GetStarted.html)

#### Step 2: Install necessary tools on your instance

1. SSH into your server
2. [Install Go](https://go.dev/doc/install)
3. Install git and make `yum install git make`

### Step 3: Set up a separate github repository for your world definitions

By separating world files from the engine, we can enable local development

1. Create a new github repository - make it private if you can
2. Ensure that the root of your project contains the contents of your `_datafiles`
3. add a `.gitignore` to avoid clobbering user files
```
world/default/users/*
!world/default/users/admin.yaml
```

### Step 4: Set up GoMUD on the remote server

Optional pre-step if you have a private github repository for your world files.
* [Create a keypair and add it to your github account](https://docs.github.com/en/authentication/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent)

1. From /home/ec2-user `git clone git@github.com:Volte6/GoMud.git`
2. If you have a separate datafiles repository `git clone git@github.com:YOURNAME/YourRepo.git _datafiles`

### Step 5: Build GoMud and launch the service for the first time

1. Build GoMud and copy the executable to `/home/ec2-user/`
```
cd /home/ec2-user/GoMud
PATH=$PATH:/usr/local/go/bin make
cp go-mud-server /home/ec2-user/go-mud-server
```
2. Create a new file `touch /home/ec2-user/go-mud-server.service`
3. Add the following contents to the file.
```
[Unit]
Description=Service to run go mud.

[Install]
WantedBy=multi-user.target

[Service]
Type=simple
ExecStart=/home/ec2-user/go-mud-server
WorkingDirectory=/home/ec2-user
Restart=always
RestartSec=5
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=%n
```
4. Add the service to the daemon `sudo systemctl daemon-reload`
5. Enable the service `sudo systemctl enable go-mud-server.service`
5. Launch the service `sudo systemctl start go-mud-server.service`


### Step 6: How to update GoMud and world files
1. Create a new script `touch /home/ec2-user/update.sh`
2. Make the script executable `chmod +x /home/ec2-user/update.sh`
3. Add the following lines to the script.
```
cd GoMud

git fetch --all
git checkout master
git pull

# Build executable
PATH=$PATH:/usr/local/go/bin
make
cp go-mud-server /home/ec2-user/go-mud-server

cd ../_datafiles
git fetch --all
git checkout <your branch>
git pull
```
4. To update, execute the following commands:
```
sudo systemctl stop go-mud-server.service
./update.sh
sudo systemctl start go-mud-server.service
```
