# LinkPreview.net Proxy Server with Caching

## Description
LinkPreview.net Proxy Server with Caching and no 3rd party dependencies written in Go.

You can use this as a proxy server for LinkPreview API requests. It will add a simple caching layer and eliminate the need to expose your API keys on the frontend.

## Installation


How to use:

1. Download official executable. [Download for Linux 64bit](https://github.com/interactive32/lpproxy/releases/download/v1.0.0/lpproxy-1.0.0.linux-amd64.zip)
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

http://your-ip-address/?q=http://google.com


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




