#! /bin/bash

installSoftware() {
    apt -qq -y install nginx default-mysql-client
    apt -qq -y -t $(lsb_release -sc)-backports install golang-go npm
}

installMyTasks() {
    curl -Lo- https://github.com/sunshineplan/mytasks/releases/download/v1.0/release.zip | tar zxC /var/www/mytasks
    cd /var/www/mytasks
    chmod +x mytasks
    ./mytasks install
}

configMyTasks() {
    read -p 'Please enter metadata server: ' server
    read -p 'Please enter VerifyHeader header: ' header
    read -p 'Please enter VerifyHeader value: ' value
    read -p 'Please enter unix socket(default: /run/mytasks.sock): ' unix
    [ -z $unix ] && unix=/var/www/mytasks/mytasks.sock
    read -p 'Please enter host(default: 127.0.0.1): ' host
    [ -z $host ] && host=127.0.0.1
    read -p 'Please enter port(default: 12345): ' port
    [ -z $port ] && port=12345
    read -p 'Please enter log path(default: /var/log/app/mytasks.log): ' log
    [ -z $log ] && log=/var/log/app/mytasks.log
    mkdir -p $(dirname $log)
    sed "s,\$server,$server," /var/www/mytasks/config.ini.default > /var/www/mytasks/config.ini
    sed -i "s/\$header/$header/" /var/www/mytasks/config.ini
    sed -i "s/\$value/$value/" /var/www/mytasks/config.ini
    sed -i "s,\$unix,$unix," /var/www/mytasks/config.ini
    sed -i "s,\$log,$log," /var/www/mytasks/config.ini
    sed -i "s/\$host/$host/" /var/www/mytasks/config.ini
    sed -i "s/\$port/$port/" /var/www/mytasks/config.ini
    service mytasks start
}

writeLogrotateScrip() {
    if [ ! -f '/etc/logrotate.d/app' ]; then
	cat >/etc/logrotate.d/app <<-EOF
		/var/log/app/*.log {
		    copytruncate
		    rotate 12
		    compress
		    delaycompress
		    missingok
		    notifempty
		}
		EOF
    fi
}

createCronTask() {
    cp -s /var/www/mytasks/scripts/mytasks.cron /etc/cron.monthly/mytasks
    chmod +x /var/www/mytasks/scripts/mytasks.cron
}

setupNGINX() {
    cp -s /var/www/mytasks/scripts/mytasks.conf /etc/nginx/conf.d
    sed -i "s/\$domain/$domain/" /var/www/mytasks/scripts/mytasks.conf
    sed -i "s,\$unix,$unix," /var/www/mytasks/scripts/mytasks.conf
    service nginx reload
}

main() {
    read -p 'Please enter domain:' domain
    installSoftware
    installMyTasks
    configMyTasks
    writeLogrotateScrip
    createCronTask
    setupNGINX
}

main
