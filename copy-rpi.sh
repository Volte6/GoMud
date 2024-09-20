
RPI_HOST="mud@mud.local"
RPI_HOST_PATH="/home/mud"
RPI_BIN="go-mud-server-rpi"

# Build the raspberry pi binary... building on the rpi is problematic.
make build_rpi

# Kill the process before overwriting the binary
ssh ${RPI_HOST} 'sudo pkill mud-server'
# Copy the binary over, delete the local binary, run the server again using the script setup ont he server
scp ./${RPI_BIN} ${RPI_HOST}:${RPI_HOST_PATH}/mud/${RPI_BIN} && \
rm ${RPI_BIN} && \
ssh ${RPI_HOST} 'cd ${RPI_HOST_PATH}; sudo ./startup-run-mud.sh'
