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

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin

app_path={APP_PATH}

im_start(){
    isStart=`ps -ef|grep 'imail start' |grep -v grep|awk '{print $2}'`
    if [ "$isStart" == '' ];then
        echo -e "Starting imail... \c"
        cd $app_path && ./imail start
            isStart=""
            while [[ "$isStart" == "" ]];
            do
                echo -e ".\c"
                sleep 0.5
                isStart=$(lsof -n -P -i:25|grep LISTEN|grep -v grep|awk '{print $2}'|xargs)
                let n+=1
                if [ $n -gt 15 ];then
                    break;
                fi
            done
            if [ "$isStart" == '' ];then
                    echo -e "\033[31mfailed\033[0m"
                    echo '------------------------------------------------------'
                    tail -n 20 ${mw_path}/logs/error.log
                    echo '------------------------------------------------------'
                    echo -e "\033[31mError: mw service startup failed.\033[0m"
                    return;
            fi
            echo -e "\033[32mdone\033[0m"
    else
            echo "Starting imail(pid $(echo $isStart)) already running"
    fi
}

im_stop(){
	echo "im_stop"
}

im_reload(){
	echo "im_reload"
}


im_status(){
	echo "im_status"
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