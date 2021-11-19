# LinkPreview.net Proxy Server with Caching

## Description
LinkPreview API/Image Proxy Server with Caching written in Go.

Features:

* In-memory caching layer for API requests
* Eliminate the need to expose your LinkPreview API keys on the frontend
* Image proxy to prevent leaking client IP addresses
* Simple Referer header protection (accept requests from your domain only)

## Installation


How to use:

1. Download official executable. 
[Download for Linux 64bit](https://github.com/interactive32/lpproxy/releases/download/v2.0.0/lpproxy-2.0.0.linux-amd64.tar.gz)
[Download for Windows 64bit](https://github.com/interactive32/lpproxy/releases/download/v2.0.0/lpproxy-2.0.0.windows-amd64.zip)

2. Create .env file configuration, see .env.example


Sample .env file:

```
ADDR="localhost:8000"
LINK_PREVIEW_KEY="123456"
```

3. Start your server with ```./lpproxy```


Note: If you want to start SSL server, add your certificate files to the .env config or create self-signed ones with:
```
openssl req -new -x509 -days 365 -nodes -out cert.pem -keyout cert.key
```

## Usage

Make sure your proxy is running, then simply use this instead of api.linkpreview.net endpoint:

```
http://your-ip-address/linkpreview/?q=http://google.com
```


Serve images through your proxy:

```
http://your-ip-address/imageproxy/?src=https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_150x54dp.png
```

For more options on image proxy, see https://github.com/willnorris/imageproxy

## Installing as a service

How to install and run linkpreview proxy as a Systemd service in Linux:

```cd /etc/systemd/system```

Create a file named lpproxy.service and include the following:

```
[Unit]
Description=LinkPreviewProxy

[Service]
User=root
WorkingDirectory=/path/where/your/lpproxy/binary/is/located/
ExecStart=/path/where/your/lpproxy/binary/is/located/lpproxy
Restart=always

[Install]
WantedBy=multi-user.target
```

Reload the service files to include the new service.
```sudo systemctl daemon-reload```

Start your proxy
```sudo systemctl start lpproxy.service```

To check the status of your proxy
```sudo systemctl status lpproxy.service```

To enable proxy on every reboot
```sudo systemctl enable lpproxy.service```


# Compile from sources

First, download and install [Golang](https://golang.org/), then clone the project and compile:
```
git clone git@github.com:interactive32/lpproxy.git
cd lpproxy
go build
```

You can also cross compile for specific architecture:
```
env GOOS=windows GOARCH=amd64 go build
```

