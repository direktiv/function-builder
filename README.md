# apps

./builder.sh generate swagger_v1.0.0.yaml

// create /tmp/app and init

./builder.sh init jens


<!-- replace github.com/direktiv/apps/go => /home/jensg/go/src/github.com/direktiv/apps/go -->

```
docker run -v `pwd`:/tmp/app builder init aws-cli
```



## Cancel Function

```
  /: 
    delete:
      parameters:
        - name: Direktiv-ActionID
          in: header
          type: string
          description: |
            On cancel Direktiv sends a DELETE request to
            the action with id in the header
      x-direktiv:
        cancel: echo 'cancel {{ .DirektivActionID }}'
      responses:
        200:
```