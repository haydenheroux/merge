<a href="https://www.merge.zone"><img width="1171" height="1401" alt="image" src="https://github.com/user-attachments/assets/4cf16273-4cea-4b1a-972c-0f761d863266" /></a>

```mermaid
architecture-beta
    group frontend(internet)[Frontend]

    service client(internet)[Vanilla HTML5] in frontend

    group backend(server)[Backend]

    service server(server)[Go Server] in backend
    service templ(server)[Templ Components] in backend

    group api(cloud)[APIs]

    service github(cloud)[GitHub API] in api

    client:L -- R:templ
    templ:B -- T:server
    server:L -- R:github
```
