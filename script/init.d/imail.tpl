#!/bin/bash
# chkconfig: 2345 55 25
# description: Imail Service

### BEGIN INIT INFO
# Provides:          bt
# Required-Start:    $all
# Required-Stop:     $all
# Default-Start:     2 3 4 5
# Default-Stop:      0 1 6
# Short-Description: starts Imail
# Description:       starts the Imail
### END INIT INFO


im_start(){

}

im_stop(){
	
}

im_reload(){
	
}


im_status(){
	
}

case "$1" in
    'start') im_start;;
    'stop') im_stop;;
    'reload') im_reload;;
    'restart') 
        im_stop
        im_start;;
    'status') im_status;;
esac