# Neomutt OAuth2 utility

## Overview

This is a program ported from Mutt/Neomutt that performs OAuth2.0 authentication to Golang.


## usage

This program call by Mutt/Neomutt `imap_oauth_refresh_command`

https://neomutt.org/guide/reference.html

##### 1. generate your oauth api 

access `https://console.cloud.google.com/welcome`

##### 2. create project

create a suitable project

##### 3. create OAuth credentials

select `API and service` - `credentials` and create `OAuth client ID`

##### 4. generate your oauth tokens

Download this program and do below command

```
./oauth \
  --user=<your account> \
  --client_id=<3. client_id> \ 
  --client_secret=<3. client_secret> \
  --generate_oauth2_token
```

Your authorization URL will be displayed, so access

```
To authorize token, visit this url and follow the directions:
  https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=<your id>&redirect_uri=http%3A%2F%2Flocalhost&response_type=code&scope=https%3A%2F%2Fmail.google.com%2F&state=XVlBzgbaiCMRAjWw
```

You will eventually be forwarded to a dummy URL and code will display the access code. Enter the display content

```
Enter verification code:
```


Get access code and refresh token

```
Refresh Token: <your refresh token>
Access Token: <your access token>
Access Token Expiration Seconds: <access token expire date> 
```

##### muttrc

Add a setting to muttrc and set it to get an access token at startup


The example puts the program in ~/.mutt/oauth

```
set imap_oauth_refresh_command="~/.mutt/oauth \
  --quiet \
  --user=<your google account> \
  --client_id=<your client_id> \
  --client_secret=<your client_secret> \
  --refresh_token=<your refresh token> 

set smtp_oauth_refresh_command="~/.mutt/oauth \
  --quiet \
  --user=<your google account> \
  --client_id=<your client_id> \
  --client_secret=<your client_secret> \
  --refresh_token=<your refresh token> 
```

