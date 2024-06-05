# authentication-service


## Description 
Basic authentication service using

## Technologies

* Go 1.20
* JWT + RT
* mongoDB


### Usage
Clone the repository:
```
git clone git@github.com:aidostt/authentication-service.git
```

Change authentication-service/.env file and add your mongoDB credentials:
```
MONGO_URI=_YOUR MONGO_URI HERE_
MONGO_USER=_YOUR MONGO_USER HERE_
MONGO_PASS=_YOUR MONGO_PASS HER_
...
```

Run a program:
```
go run ./cmd/api/.
```



### Documentation


Available endpoints and it's methods: 
```
GET    /ping                     --> authentication-service/internal/handlers.(*Handler).Init.func1 (6 handlers)
POST   /api/users/sign-up        --> authentication-service/internal/handlers.(*Handler).userSignUp-fm (6 handlers)
POST   /api/users/sign-in        --> authentication-service/internal/handlers.(*Handler).userSignIn-fm (6 handlers)
POST   /api/users/auth/refresh   --> authentication-service/internal/handlers.(*Handler).userRefresh-fm (6 handlers)
GET    /api/users/healthcheck    --> authentication-service/internal/handlers.(*Handler).healthcheck-fm (7 handlers)
POST   /api/users/sign-out       --> authentication-service/internal/handlers.(*Handler).logout-fm (6 handlers)
```


### To run the container you should write:
docker build -t authentication-service .
docker run -p 5050:5050 --env-file .env -ti authentication-service

