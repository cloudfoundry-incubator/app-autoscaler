<link href="https://raw.github.com/clownfart/Markdown-CSS/master/markdown.css" rel="stylesheet"></link>
[![Build Status](https://runtime-og.ci.cf-app.com/api/v1/pipelines/autoscaler/jobs/unit-tests/badge?ts=1)](https://runtime-og.ci.cf-app.com/pipelines/autoscaler)

# App-AutoScaler

This is an incubation project for Cloud Foundry. You can follow the development progress on [Pivotal Tracker][t].

The `App-AutoScaler` provides the capability to adjust the computation resources for Cloud Foundry applications through

* Dynamic scaling based on application performance metrics
* Scheduled scaling based on time

The `App-AutoScaler` is provided as a Cloud Foundry service offering. Any application bound with `App-AutoScaler` service will be able to use it. It has the following components:

* `api` : provides public APIs to manage scaling policy
* `servicebroker`: implements the [Cloud Foundry service broker API][k]
* `metricscollector`: collects container's memory usage
* `eventgenerator`: aggreates memory metrics, evaluates scaling rules and triggers events for dynamic scaling
* `scheduler`: manages the schedules in scaling policy and trigger events for scheduled scaling
* `scalingengine`: takes the scaling actions based on dynamic scaling rules or schedules


## Development

### System requirements

* Java 8 or above
* [Apache Maven][b] 3
* Node 6.2 or above
* NPM 3.9.5 or above
* [Cloud Foundry cf command line] [f]
* Go 1.7

### Database requirement

The `App-AutoScaler` uses Postgres as the backend data store. To download and install, refer to [PostgreSQL][p] web site.


### Setup

To set up the development, firstly clone this project

```shell
$ git clone https://github.com/cloudfoundry-incubator/app-autoscaler.git
$ cd app-autoscaler
$ git submodule update --init --recursive
```


#### Initialize the Database

```shell
createuser postgres -s
psql postgres://postgres@127.0.0.1:5432 -c 'DROP DATABASE IF EXISTS autoscaler'
psql postgres://postgres@127.0.0.1:5432 -c 'CREATE DATABASE autoscaler'

mvn package
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=api/db/api.db.changelog.yml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=servicebroker/db/servicebroker.db.changelog.json update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=scheduler/db/scheduler.changelog-master.yaml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=scheduler/db/quartz.changelog-master.yaml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=src/autoscaler/metricscollector/db/metricscollector.db.changelog.yml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=src/autoscaler/eventgenerator/db/dataaggregator.db.changelog.yml update
java -cp 'db/target/lib/*' liquibase.integration.commandline.Main --url jdbc:postgresql://127.0.0.1/autoscaler --driver=org.postgresql.Driver --changeLogFile=src/autoscaler/scalingengine/db/scalingengine.db.changelog.yml update
```

#### Generate TLS Certificates

```shell
./scripts/generate_test_certs.sh
```


### Unit tests

```shell
pushd api
npm install
npm test
popd

pushd servicebroker
npm install
npm test
popd

go install github.com/onsi/ginkgo/ginkgo
export DBURL=postgres://postgres@localhost/autoscaler?sslmode=disable
pushd src/autoscaler
ginkgo -r -race -randomizeAllSpecs
popd

pushd scheduler
mvn test
popd
```

### Integration tests

```shell
pushd api
npm install
popd

pushd servicebroker
npm install
popd

pushd scheduler
mvn package -DskipTests
popd

go install github.com/onsi/ginkgo/ginkgo
export DBURL=postgres://postgres@localhost/autoscaler?sslmode=disable
ginkgo -r -race -p -randomizeAllSpecs src/integration
```

## Deploy

Go to [app-autoscaler-release][r] project to BOSH deploy `App-AutoScaler` 

## Register service broker
You can register `App-AutoScaler` with command:

```shell
cf create-service-broker App-AutoScaler <brokerUserName> <brokerPassword> <brokerURI>
cf enable-service-access autoscaler
```

## Use `App-AutoScaler` 

Now, you can play with `App-AutoScaler`.

Firstly create a `App-AutoScaler` service, and bind to you application

``` shell
cf create-service autoscaler  autoscaler-free-plan  <service_instance_name>
cf bind-service <app> <service_instance_name> -c <policy>
```


## License

This project is released under version 2.0 of the [Apache License][l].


[b]: https://maven.apache.org/
[c]: http://couchdb.apache.org/
[d]: http://www.eclipse.org/m2e/
[e]: http://www.cloudant.com
[f]: https://github.com/cloudfoundry/cli/releases
[k]: http://docs.cloudfoundry.org/services/api.html
[l]: LICENSE
[t]: https://www.pivotaltracker.com/projects/1566795
[p]: https://www.postgresql.org/
[r]: https://github.com/cloudfoundry-incubator/app-autoscaler-release
