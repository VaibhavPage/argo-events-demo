# Event-Driven Parameterized Jupyter Notebooks

Jupyter notebooks are prevalent in data science community to develop models, run analysis and generate reports, etc.
But in many situations, a data scientist must feed varying parameters to tune the notebook to generate an optimal model. 
Tools like [papermill](https://papermill.readthedocs.io/en/latest/) makes it easy to parameterize the notebook and Argo Events makes it super easy
to set up **event-driven** parameterized notebooks.

# Prerequisites
1. Install [Argo Workflows](https://github.com/argoproj/argo/blob/master/docs/getting-started.md).
2. Install [Argo Events](https://argoproj.github.io/argo-events/installation/).
3. Install NATS,
    
        kubectl -n argo-events apply -f https://raw.githubusercontent.com/VaibhavPage/argo-events-demo/master/nats-deploy.yaml

4. Install Minio,

        kubectl -n argo-events apply -f https://raw.githubusercontent.com/VaibhavPage/argo-events-demo/master/artifact-minio.yaml

        kubectl -n argo-events apply -f https://raw.githubusercontent.com/VaibhavPage/argo-events-demo/master/minio-deploy.yaml

# Setup
In this demo, we are going to set up an image processing pipeline using 2 notebooks. Lets consider the ArgoProj icon image

<br/>
<br/>

<p align="center">
  <img src="https://github.com/argoproj/argo-events/blob/master/docs/assets/argo.png?raw=true" alt="Argo"/>
</p>

<br/>
<br/>

1. The [first notebook](https://github.com/VaibhavPage/argo-events-demo/blob/master/img-out.ipynb) will take the clean ArgoProj logo and add noise to it.
2. The [second notebook](https://github.com/VaibhavPage/argo-events-demo/blob/master/matcher.ipynb) is going to determine the similarity between clean image and the image with noise.
   If the match is > 80%, then model is optimal, else we need to tune the noise parameters.

<br/>

<p align="center">
  <img src="https://github.com/VaibhavPage/argo-events-demo/blob/master/out/argo-demo.png?raw=true" alt="Argo Demo"/>
</p>

<br/>

3. We will set up two gateways, Webhook and Minio. The webhook gateway will listen to HTTP requests to tune the 
   notebook to add noise to image. The notebook will store the noisy image to Minio.
   
4. The minio gateway will listen to file drop events for a specific bucket. Once the noisy image is dropped into that bucket,
   we will run the second notebook that determines the similarity of images.
   
<br/>

<p align="center">
  <img src="https://github.com/VaibhavPage/argo-events-demo/blob/master/argo-events-demo-pipeline.png?raw=true" alt="Argo Demo"/>
</p>

<br/>

5. Create webhook event source. It consist configuration for gateway to listen for HTTP POST requests on port 12000.


        kubectl -n argo-events apply -f https://raw.githubusercontent.com/VaibhavPage/argo-events-demo/master/webhook-event-source.yaml

6. Create webhook gateway,

        kubectl -n argo-events apply -f https://raw.githubusercontent.com/VaibhavPage/argo-events-demo/master/webhook-gateway.yaml
        
7. Create webhook sensor,

        kubectl -n argo-events apply -f https://raw.githubusercontent.com/VaibhavPage/argo-events-demo/master/webhook-sensor.yaml

8. Lets inspect webhook sensor,

        apiVersion: argoproj.io/v1alpha1
        kind: Sensor
        metadata:
          name: webhook-sensor
          labels:
            sensors.argoproj.io/sensor-controller-instanceid: argo-events
        spec:
          template:
            spec:
              containers:
                - name: sensor
                  image: argoproj/sensor:v0.13.0-rc
                  imagePullPolicy: Always
              serviceAccountName: argo-events-sa
          dependencies:
            - name: test-dep
              gatewayName: webhook-gateway
              eventName: example
          subscription:
            http:
              port: 9300
          triggers:
            - template:
                name: webhook-workflow-trigger
                k8s:
                  group: argoproj.io
                  version: v1alpha1
                  resource: workflows
                  operation: create
                  source:
                    resource:
                      apiVersion: argoproj.io/v1alpha1
                      kind: Workflow
                      metadata:
                        generateName: webhook-
                      spec:
                        entrypoint: noisy
                        arguments:
                          parameters:
                            - name: filterA
                              value: "5"
                            - name: filterB
                              value: "5"
                            - name: sVSp
                              value: "0.5"
                            - name: amount
                              value: "0.004"
                        templates:
                          - name: noisy
                            serviceAccountName: argo-events-sa
                            inputs:
                              parameters:
                                - name: filterA
                                - name: filterB
                                - name: sVSp
                                - name: amount
                            container:
                              image: metalgearsolid/demo-blur-argo-logo:latest
                              command: [papermill]
                              imagePullPolicy: Always
                              env:
                                - name: AWS_ACCESS_KEY_ID
                                  value: minio
                                - name: AWS_SECRET_ACCESS_KEY
                                  value: minio123
                                - name: AWS_DEFAULT_REGION
                                  value: us-east-1
                                - name: BOTO3_ENDPOINT_URL
                                  value: http://minio-service.argo-events.svc:9000
                              args:
                                - "noise.ipynb"
                                - "s3://output/noisy-out.ipynb"
                                - "-p"
                                - "filterA"
                                - "{{inputs.parameters.filterA}}"
                                - "-p"
                                - "filterB"
                                - "{{inputs.parameters.filterB}}"
                                - "-p"
                                - "sVSp"
                                - "{{inputs.parameters.sVSp}}"
                                - "-p"
                                - "amount"
                                - "{{inputs.parameters.amount}}"
                  parameters:
                    - src:
                        dependencyName: test-dep
                        dataKey: body.filterA
                      dest: spec.arguments.parameters.0.value
                    - src:
                        dependencyName: test-dep
                        dataKey: body.filterB
                      dest: spec.arguments.parameters.1.value
                    - src:
                        dependencyName: test-dep
                        dataKey: body.sVSp
                      dest: spec.arguments.parameters.2.value
                    - src:
                        dependencyName: test-dep
                        dataKey: body.amount
                      dest: spec.arguments.parameters.3.value

1. 


