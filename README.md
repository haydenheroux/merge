<a href="https://www.merge.zone"><img width="960" height="930" alt="image" src="https://github.com/user-attachments/assets/94c85dd6-2cb2-4629-bed2-f8ce793f1a7c" /></a>

```mermaid
architecture-beta
    group api(cloud)[API]

    service server(server)[Go Server] in api

    group frontend(internet)[Frontend]

    service client(internet)[HTML Pages] in frontend

    server{group}:L -- R:client{group}
```
