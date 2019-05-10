# Project Layout
https://github.com/golang-standards/project-layout

# AWS
https://docs.aws.amazon.com/sdk-for-go/api/service/s3/

# Vendoring
https://golang.github.io/dep/

# Presigned POST S3
https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html

https://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-authentication-HTTPPOST.html
https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingHTTPPOST.html
Note

Query string authentication is not supported for POST.


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

````bash
https://assertuploader.s3.eu-west-1.amazonaws.com/a79b143e-6218-450b-a8e8-18d00d788b8b?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=AKIASWEEC46WNIHR44WH%2F20190510%2Feu-west-1%2Fs3%2Faws4_request&X-Amz-Date=20190510T204051Z&X-Amz-Expires=900&X-Amz-SignedHeaders=host&X-Amz-Signature=3df259f4cacbf54a157673c67b285b71ff28ae3d01df52b59d203c9af01fba59
````

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

# S3 performance
https://aws.amazon.com/blogs/aws/amazon-s3-performance-tips-tricks-seattle-hiring-event/
https://docs.aws.amazon.com/AmazonS3/latest/dev/request-rate-perf-considerations.html
Superseded by ->
https://aws.amazon.com/about-aws/whats-new/2018/07/amazon-s3-announces-increased-request-rate-performance/

https://dzone.com/articles/how-to-optimize-amazon-s3-performance


