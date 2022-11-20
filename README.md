# sysalert
# Simple sys stats alert tool for sending emails when your system exceeds configured thresholds


# Install
```
go install
cp sample/sysalert.yaml .
```

## Run on cronicle scheduler every 5 minutes
```
wget -c https://github.com/jshiv/cronicle/releases/download/v0.3.8/cronicle_0.3.8_Linux_x86_64.tar.gz -O - | tar -xz
sudo mv cronicle /usr/local/bin/cronicle
cronicle run --command "./go/bin/sysalert" --cron "@every 5m"
```