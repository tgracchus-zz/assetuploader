# Asset uploader

## How to
### How to create distributions
```bash
build/distribution.sh
```
In order to run them:  
* Make sure to define  
export AWS_ACCESS_KEY_ID=XXXXX  
export AWS_SECRET_ACCESS_KEY=XXXXX  
* Run
```bash
./assetuploader-1.0-darwin-x86_64 --region=${AWS_REGION} --bucket=${AWS_BUCKET}
```

### How to run in-place
* Make sure to define:  
export AWS_ACCESS_KEY_ID=XXXXX  
export AWS_SECRET_ACCESS_KEY=XXXXXX  
export AWS_REGION=XXXXX  
export AWS_BUCKET=XXXXX  
Otherwise the scrip will fail.  
* Run
```bash
build/run.sh
```

### How to run with docker
* Make sure to define:  
export AWS_ACCESS_KEY_ID=XXXXX  
export AWS_SECRET_ACCESS_KEY=XXXXXX  
export AWS_REGION=XXXXX  
export AWS_BUCKET=XXXXX  
Otherwise the scrip will fail.  
* Run
```bash
build/distribution.sh
docker run -e AWS_ACCESS_KEY_ID="XXXXXX" -e AWS_SECRET_ACCESS_KEY="XXXXXX" -e AWS_REGION="eu-west-1" -e AWS_BUCKET="assetuploader-1" -it -p 8080:8080 assetuploader:1.0
```


### How to run Test
* Run  
```bash
build/test.sh
```

### How to run Integration Test
* Make sure to define:    
export AWS_ACCESS_KEY_ID=XXXXX  
export AWS_SECRET_ACCESS_KEY=XXXXXX  
export AWS_REGION=XXXXX  
export AWS_BUCKET=XXXXX  
Otherwise the scrip will fail.
* Run  
```bash
build/itest.sh
```

## S3 file schema
Before explaining the actual endpoints it´s worth explaining the s3chema used in the app.  

bucket:    
  -> temp/{assetID}  
  -> uploaded/{assetIDD}  

Theres two reasons for this schema:
1. Prevent the user to use the presigned put url for a get before it´s marked as uploaded, since the only difference between both requests is the method.  
2. Prevent the user to overwrite the files marked as uploaded

So, this is the usual flow:
* Post /assets/{assetID} => placeholder file is created inside uploaded/{assetID} with url expiration time
  The presigned url points to /temp, so the file will be uploaded to /temp.  

* Put /assets/{assetID} =>  placeholder file is check for uploaded status.
  * If uploaded => trow an error, already uploaded
  * If not uploaded => schedule a job, to be executed after url expiration, using the url expiration time in the placeholder. That job will copy the temp/{assetID} to uploaded/{assetID} and update tags to mark it as uploaded.
  
* When get /assets/{assetID} is called the metata file is check for uploaded status.
  * If uploaded: generate the presigned url pointing to the uploaded/{assetID} file
  * If not uploaded: trow an error


Note:  
S3 paths are no longer affected by prefixes.
[No prefix considerations for s3 anymore](https://aws.amazon.com/about-aws/whats-new/2018/07/amazon-s3-announces-increased-request-rate-performance/)

## Endpoints
### POST ​​/asset  
* **Description**:  
Creates a new asset with a random uuid and returns a url to put the asset in s3.
It also adds a placeholder to final destination
* **Body**:  
empty  

* **Response**:  
```
{
"upload_url":​​"<s3-signed-url-for-upload>", ​​"id":​​"<asset-id>"
}
```
Response code | Description
------------ | -------------
201 | Asset id created
500 | Internal Error  
  
* **Technical Notes**:  

  * **PUT ​​/asset/<asset-id>**   
Since the 
[S3 consistency model](https://docs.aws.amazon.com/AmazonS3/latest/dev/Introduction.html#ConsistencyModel
)
supports read-after-write consistency, that is (only): put and get same object.  
When the PUT /asset/assetID checks that placeholder, it should be ok to read.  
However, a malicious user can still try to auto generate uuids and query the api so we lose the 
read-after-write consistency, because we will have a write-after-read -> eventual consistency.
 
  * **POST to s3:**  
  The original problem statement ask for a url which can be used by a post directly to s3.
  Even a post query can be made to s3, it´s intended for [browser upload.](https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-authentication-HTTPPOST.html)
   
    So, when trying to do a post to a bucket/assetID with query params auth, it fails:
  ```bash
     https://assertuploader.s3.eu-west-1.amazonaws.com/a79b143e-6218-450b-a8e8-18d00d788b8b?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIASWEEC46WNIHR44WH%2F20190510%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Date=20190510T204051Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=3df259f4cacbf54a157673c67b285b71ff28ae3d01df52b59d203c9af01fba59
  ```
  
   ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <Error>
        <Code>MethodNotAllowed</Code>
        <Message>The specified method is not allowed against this resource.</Message>
        <Method>POST</Method>
        <ResourceType>OBJECT</ResourceType>
        <RequestId>A6BC66C6E743C2BA</RequestId>
        <HostId>jvYsrZh9D3cdDPgoVe8MuiVunH5HfPIiWU0L8MW8NAUG462w/YHiG1reg4OrMNjowYBX5gPvOgA=</HostId>
    </Error>
   ```
  
  Similarly, query string authentication is not supported for POST. 
  ````bash
  https://assertuploader.s3.eu-west-1.amazonaws.com/?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIASWEEC46WNIHR44WH%2F20190510%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Date=20190510T201100Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=35b37b840d2e5a68f0716fa66aef10405ad788311367b2fe82b9b7baa133552a
  ````
  ```xml
  <?xml version="1.0" encoding="UTF-8"?>
  <Error>
      <Code>InvalidArgument</Code>
      <Message>Query String Parameters not allowed on POST requests.</Message>
      <ArgumentName>X-Amz-SignedHeaders</ArgumentName>
      <ArgumentValue>null</ArgumentValue>
      <RequestId>B15BD091C3F46199</RequestId>
      <HostId>VEyCIjFNnAB94JgiTXQtInRkVyXFbe4d5QwY80wctbrh5UYzNSOo8WEWRYo2trA1m0j0LIToCvg=</HostId>
  </Error>
  ````
  
  So, [we need to use a put](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-presigned-urls.html), not a post, with the url returned

  * **POST correctness:**  
    As explained [here](https://stackoverflow.com/questions/630453/put-vs-post-in-rest), a post is more intended when we don´t know the id of the new resource and a put when we know it.  
    In our case, the asset id is generated by the POST /asset endpoint. So, conceptually, to me, a put to s3 seems more adecuated than a post
     

### PUT ​​/asset/<asset-id>  
* **Description:** 
Will mark the upload operation as completed.

* **Body:** 
```
{ ​​​"Status":​​"uploaded" }
```  

* **Response:**  
Empty

Response code | Description
------------ | -------------
202 | Query accepted
400 | If the request is incorrect
404 | If the asset id is not found
500 | Internal Error  


* **Technical Notes**:  
This is the tricky endpoint, since the user can put the asset again and again as long as the signed url is valid.  
That esentially introduces the problem of marking as uploaded the correct file and make sure it remains the same.
For this reason, I can foresee 4 approaches:  
  1. Do it before expiration of the put request:  
  Imagine:  
    -> Post file1 to asset1-> Put /asset/asset1  
    then, at the same time:  
    -> Post file2 to asset1 -> Put /asset/asset1    </p>
    Given than s3 operations are atomic at file level:  
    Queries will compete against each other and we will end up with a file marked as complete, 
    but we are uncertain about if it is file1 or file2.   
    So, the consistency model is weak in this case, not even eventual consistency

  2. Do not allow the operation until the post is expired:  
Hard limitation on the api, simple and effective solution which leads to a more strong consistency model.  
It´s in fact still eventual consistency since tagging in s3 is eventual consistent (tags are used to mark an object as uploaded)  
It also it delegates the problem since it keep the client trying to mark the file as uploaded

  3. Try to get a lock in the object, using a distributed lock for updating it. This solution provides a more consistent 
  solution but it prevents to user to progress if the request fail to get the lock.

  4. Schedule a task to mark the asset as uploaded after the pregisned put url is expired:
  Introduces the need for a queue or some kind of scheduling. Eventual consistency model, but, to me it provides the nicest user experience, so: **I decided to go for the last approach.**

### GET ​​/asset/<asset-id>  
* **Description:**   
Will get a signed s3 url for getting the object

* **Response:**  
```
{ ​​​"Download_url":​​"<s3-signed-url-for-upload>" } 
```

Response code | Description
------------ | -------------
200 | Query succeed
400 | If the request is incorrect
404 | If the asset id is not found
500 | Internal Error


* **Technical Notes:**  
In the PUT ​​/asset/<asset-id> endpoint we mark the asset as completed. 
This is done with a tag in the uploaded object. Since tags are eventually consistent,
it might take a bit longer after the object is marked as uploaded. But, since the ### PUT ​​/asset/<asset-id> 
is async, it does not matter.


## Solution concurreny model.
- Post, S3 Put, Put Completed => ok  
- Post, S3 Put, Put Completed,Put Completed => ok => Put is idempotent
- Post, S3 Put, Put Completed, S3 Put, Put Completed => ok => Will pick second s3 put given s3 converged (its eventual consistent in this case)  
- Post, S3 Put, Put Completed, S3 Put => => ok => Will pick second s3 put given s3 converged (its eventual consistent in this case)  

## Bonus point
- Create a persistent SimpleScheduler, so if the instance goes down, we can still be able to resume the jobs. 


## Project Layout
Following pkg, cmd and build patterns as seen here
https://github.com/golang-standards/project-layout

## Vendoring
Dep is used for vendoring
https://golang.github.io/dep/


## AWS SDK
https://docs.aws.amazon.com/sdk-for-go/api/service/s3/











https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/
https://stackoverflow.com/questions/42564058/how-to-use-local-docker-images-with-minikube

helm install --name="assetuploader-1.0.0" build/install/assetuploader
helm delete assetuploader-1.0.0 

helm del --purge assetuploader-1.0.0