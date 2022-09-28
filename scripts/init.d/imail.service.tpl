[Unit]
Description=Simple Mail Server
After=network.service
After=syslog.target

[Service]
User=imail
Group=imail
Type=simple
WorkingDirectory={APP_PATH}
ExecStart=imail service
ExecReload=/bin/kill -USR2 $MAINPID
PermissionsStartOnly=true
LimitNOFILE=5000
Restart=on-failure
RestartSec=10
RestartPreventExitStatus=1
PrivateTmp=false


[Install]
WantedBy=multi-user.target