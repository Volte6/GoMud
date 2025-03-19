### Setting Up a Raspberry Pi Zero 2 W for GoMUD

If you plan to host GoMUD on a Raspberry Pi Zero 2 W, follow these steps to prepare your device:


#### Step 1: Install Raspberry Pi OS

1. Download Raspberry Pi OS Lite from the [official Raspberry Pi website](https://www.raspberrypi.com/software/).
2. Flash the OS onto a microSD card using software like [Raspberry Pi Imager](https://www.raspberrypi.com/software/) or [Balena Etcher](https://www.balena.io/etcher/).
3. Insert the microSD card into your Raspberry Pi Zero 2 W and power it on.

#### Step 2: Set Up Networking
* NOTE: This step is not necessary if you preconfigured your wi-fi in the Pi Imager software.

1. Connect your Raspberry Pi to your Wi-Fi network by editing the `wpa_supplicant.conf` file on the boot partition of the microSD card.
2. Include the following configuration:
   ```
   country=US
   ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
   update_config=1

   network={
       ssid="YourWiFiSSID"
       psk="YourWiFiPassword"
   }
   ```
   Replace `YourWiFiSSID` and `YourWiFiPassword` with your network credentials.

#### Step 3: Install Go

1. SSH into your Raspberry Pi using its IP address:
   ```
   ssh pi@<IP_ADDRESS>
   ```
2. Update the system:
   ```
   sudo apt update && sudo apt upgrade -y
   ```
3. Install Go (GoLang):
   ```
    mkdir ~/src && cd ~/src
    sudo wget https://go.dev/dl/go1.23.4.linux-armv6l.tar.gz
    sudo tar -C /usr/local -xzf go1.23.4.linux-armv6l.tar.gz
    
    sudo nano ~/.profile
    
    ADD 2 LINES:
    PATH=$PATH:/usr/local/go/bin
	GOPATH=$HOME/go
    Exit and Save
    
    Verify GO version with: 
    go version
    

   ```
  4. Install git
  ``
   sudo apt install git -y
	``

Your Raspberry Pi Zero 2 W is now ready to host GoMUD. Proceed to the next section to set up GoMUD itself.

---

### Setting Up GoMUD

#### Step 1: Download and Install

1. Visit the official GoMUD GitHub repository: [GoMUD GitHub](https://github.com/Volte6/GoMud)[.](https://github.com/Volte6/GoMud)
2. Clone the repository using Git or download the ZIP file. Cloning is preferable for beginners as it allows you to easily pull updates and contribute back to the project.
   ```bash
   git clone https://github.com/Volte6/GoMud.git
   ```
3. Navigate to the project directory and build the application using the provided build instructions.
   ```bash
   cd GoMud
   go build
   ```

#### Step 2: Configure GoMUD

1. Locate the config.yaml under the _datafiles directory.
2. Edit settings like port, admin credentials, and logging preferences:
``
sudo nano config.yaml``
Exit and save this file.

Note: Raspberry Pi builds may need the Web Server Port to be changed from 80 to something like 8080 in order for the web front end to work. 

#### Step 3: Start the Server

Run the executable to launch the server:

```bash
cd ~/GoMud
go run .
```

The server will start on the specified port (default: 33333).

---

### Running GoMUD and Logging in as Admin

1. Open the GoMUD built-in web client by visiting the following from a separate web browser:
   ```
   http://[raspi IP address]/webclient
   ```
   For example http://192.168.50.106:8080/webclient
2. Log in with the default admin credentials set in `config.yaml`. Example:
   ```
   Username: admin
   Password: password
   ```

---

### Changing Your Password from the Default

1. Once logged in as admin, simply type the command:
   ```
   password
   ```
   Follow the prompts to set your new password.

---

### Maintaining and Restarting the Server

1\. To stop the running server for maintenance or to restart it, press `Ctrl + C` in the terminal where the server is running. This will safely terminate the process.

2\. Type ''go run .'' again to restart the server.

---
### Starting your own empty MUD world
Now that you've got the basics down, it's time to start a fresh world and begin your creation journey. 
The ``_datafiles/world`` folder has a folder called ``empty`` inside of it. Make a copy of that folder and give it your own name (i.e. ``sudo cp -r empty/ myworld/``).
Then edit your ``_datafiles/config.yaml`` ``DataFiles:`` field to point to your new world folder.

## Updating from master branch
To update your local GoMUD installation when new updates are available on the master branch of the GitHub repository:

	cd GoMud

	git pull origin master

	go build

This will fetch the latest updates and rebuild the application.

---

### Other Tips

\- Join the GoMUD community forums or Discord for tips and inspiration:
https://github.com/Volte6/GoMud/discussions
https://discord.gg/TqeM85QFdJ

### Problems?

Sometimes a raspberry pi may struggle to compile the binary directly. There are configurations changes you can make resource-wise to your raspberry pi that can solve this, but it is easier/recommended in this situation that you compile the binary locally and then copy it over to the raspberry pi.

There is a convenient `make` command to compile the pi chipset provided:

`make build_rpi` ( this will output a binary named: `go-mud-server-rpi` )

Or (window user?) just use the build comand directly:

`env GOOS=linux GOARCH=arm GOARM=5 go build -o go-mud-server-rpi`

Then you can copy the file over to your raspberry PI via SCP:

`scp ./go-mud-server-rpi pi@raspberrypi.local:/home/pi/GoMud/go-mud-server-rpi`

_Note: You may have to adjust the username/host/path information above to whatever your setup is._

