# RG - Previsão incidentes - UFU

## 🧰 Configuration

To install golang just follow the steps from website:
- https://golang.org/doc/install

To install docker and docker-compose just follow the steps from website:
- https://docs.docker.com/engine/install/
- https://docs.docker.com/compose/install/

## 🛠 How to use

Start postgres:
``` powershell
make config-up-as-compose
```

Or start postgres as container:
``` powershell
make config-up-as-container
```

Run application:
``` powershell
make run-mscproject
```

Or build application:
``` powershell
make mscproject-build
```

And then run application:
``` powershell
make run-back-mscproject
```

To stop application:
``` powershell
make stop-back-mscproject
```

To stop postgres:
``` powershell
make config-down
```

To run lint:
``` powershell
make lint
```

## 🚀 Endpoints

##### POST   `/ha-enus`                  --> darkWeb/app/msc/handlers.HAenUSHandler.post-fm
##### GET    `/ha-enus`                  --> darkWeb/app/msc/handlers.HAenUSHandler.get-fm
##### GET    `/ha-enus/JSON`             --> darkWeb/app/msc/handlers.HAenUSHandler.getJSON-fm 
##### POST   `/ha-ptbr`                  --> darkWeb/app/msc/handlers.HAptBRHandler.post-fm
##### GET    `/ha-ptbr`                  --> darkWeb/app/msc/handlers.HAptBRHandler.get-fm 
##### GET    `/ha-ptbr/JSON`             --> darkWeb/app/msc/handlers.HAptBRHandler.getJSON-fm 
##### POST   `/deep-ptbr`                --> darkWeb/app/msc/handlers.DeepptBRHandler.post-fm 
##### GET    `/deep-ptbr`                --> darkWeb/app/msc/handlers.DeepptBRHandler.get-fm 
##### GET    `/deep-ptbr/JSON`           --> darkWeb/app/msc/handlers.DeepptBRHandler.getJSON-fm 
##### POST   `/deep-enus`                --> darkWeb/app/msc/handlers.DeepenUSHandler.post-fm
##### GET    `/deep-enus`                --> darkWeb/app/msc/handlers.DeepenUSHandler.get-fm 
##### GET    `/deep-enus/JSON`           --> darkWeb/app/msc/handlers.DeepenUSHandler.getJSON-fm
##### POST   `/deep-eses`                --> darkWeb/app/msc/handlers.DeepesESHandler.post-fm 
##### GET    `/deep-eses`                --> darkWeb/app/msc/handlers.DeepesESHandler.get-fm
##### GET    `/deep-eses/JSON`           --> darkWeb/app/msc/handlers.DeepesESHandler.getJSON-fm
##### POST   `/raddle`                   --> darkWeb/app/msc/handlers.RaddleHandler.post-fm
##### GET    `/raddle`                   --> darkWeb/app/msc/handlers.RaddleHandler.get-fm
##### GET    `/raddle/JSON`              --> darkWeb/app/msc/handlers.RaddleHandler.getJSON-fm
