// The following code is AUTO-GENERATED. Please DO NOT edit.
// To update this generated code, run the following command:
// in the /codegenerator/model subdirectory of this project,
// making sure that `${GOPATH}/bin` is in your `PATH`:
//
// go install && go generate
//
// This package was generated from the schema defined at
// http://references.taskcluster.net/queue/v1/exchanges.json

// The queue, typically available at `queue.taskcluster.net`, is responsible
// for accepting tasks and track their state as they are executed by
// workers. In order ensure they are eventually resolved.
//
// This document describes AMQP exchanges offered by the queue, which allows
// third-party listeners to monitor tasks as they progress to resolution.
// These exchanges targets the following audience:
//  * Schedulers, who takes action after tasks are completed,
//  * Workers, who wants to listen for new or canceled tasks (optional),
//  * Tools, that wants to update their view as task progress.
//
// You'll notice that all the exchanges in the document shares the same
// routing key pattern. This makes it very easy to bind to all messages
// about a certain kind tasks.
//
// **Task-graphs**, if the task-graph scheduler, documented elsewhere, is
// used to schedule a task-graph, the task submitted will have their
// `schedulerId` set to `'task-graph-scheduler'`, and their `taskGroupId` to
// the `taskGraphId` as given to the task-graph scheduler. This is useful if
// you wish to listen for all messages in a specific task-graph.
//
// **Task specific routes**, a task can define a task specific route using
// the `task.routes` property. See task creation documentation for details
// on permissions required to provide task specific routes. If a task has
// the entry `'notify.by-email'` in as task specific route defined in
// `task.routes` all messages about this task will be CC'ed with the
// routing-key `'route.notify.by-email'`.
//
// These routes will always be prefixed `route.`, so that cannot interfere
// with the _primary_ routing key as documented here. Notice that the
// _primary_ routing key is alwasys prefixed `primary.`. This is ensured
// in the routing key reference, so API clients will do this automatically.
//
// Please, note that the way RabbitMQ works, the message will only arrive
// in your queue once, even though you may have bound to the exchange with
// multiple routing key patterns that matches more of the CC'ed routing
// routing keys.
//
// **Delivery guarantees**, most operations on the queue are idempotent,
// which means that if repeated with the same arguments then the requests
// will ensure completion of the operation and return the same response.
// This is useful if the server crashes or the TCP connection breaks, but
// when re-executing an idempotent operation, the queue will also resend
// any related AMQP messages. Hence, messages may be repeated.
//
// This shouldn't be much of a problem, as the best you can achieve using
// confirm messages with AMQP is at-least-once delivery semantics. Hence,
// this only prevents you from obtaining at-most-once delivery semantics.
//
// **Remark**, some message generated by timeouts maybe dropped if the
// server crashes at wrong time. Ideally, we'll address this in the
// future. For now we suggest you ignore this corner case, and notify us
// if this corner case is of concern to you.
//
// See: http://docs.taskcluster.net/queue/exchanges
//
// How to use this package
//
// This package is designed to sit on top of http://godoc.org/github.com/taskcluster/pulse-go/pulse. Please read
// the pulse package overview to get an understanding of how the pulse client is implemented in go.
//
// This package provides two things in addition to the basic pulse package: structured types for unmarshaling
// pulse message bodies into, and custom Binding interfaces, for defining the fixed strings for task cluster
// exchange names, and routing keys as structured types.
//
// For example, when specifying a binding, rather than using:
//
//  pulse.Bind(
//  	"*.*.*.*.*.*.gaia.#",
//  	"exchange/taskcluster-queue/v1/task-defined")
//
// You can rather use:
//
//  queueevents.TaskDefined{WorkerType: "gaia"}
//
// In addition, this means that you will also get objects in your callback method like *queueevents.TaskDefinedMessage
// rather than just interface{}.
package queueevents

import (
	"errors"
	"reflect"
	"strings"
	"time"
)

// When a task is created or just defined a message is posted to this
// exchange.
//
// This message exchange is mainly useful when tasks are scheduled by a
// scheduler that uses `defineTask` as this does not make the task
// `pending`. Thus, no `taskPending` message is published.
// Please, note that messages are also published on this exchange if defined
// using `createTask`.
//
// See http://docs.taskcluster.net/queue/exchanges/#taskDefined
type TaskDefined struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding TaskDefined) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding TaskDefined) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/task-defined"
}

func (binding TaskDefined) NewPayloadObject() interface{} {
	return new(TaskDefinedMessage)
}

// When a task becomes `pending` a message is posted to this exchange.
//
// This is useful for workers who doesn't want to constantly poll the queue
// for new tasks. The queue will also be authority for task states and
// claims. But using this exchange workers should be able to distribute work
// efficiently and they would be able to reduce their polling interval
// significantly without affecting general responsiveness.
//
// See http://docs.taskcluster.net/queue/exchanges/#taskPending
type TaskPending struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding TaskPending) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding TaskPending) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/task-pending"
}

func (binding TaskPending) NewPayloadObject() interface{} {
	return new(TaskPendingMessage)
}

// Whenever a task is claimed by a worker, a run is started on the worker,
// and a message is posted on this exchange.
//
// See http://docs.taskcluster.net/queue/exchanges/#taskRunning
type TaskRunning struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding TaskRunning) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding TaskRunning) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/task-running"
}

func (binding TaskRunning) NewPayloadObject() interface{} {
	return new(TaskRunningMessage)
}

// Whenever the `createArtifact` end-point is called, the queue will create
// a record of the artifact and post a message on this exchange. All of this
// happens before the queue returns a signed URL for the caller to upload
// the actual artifact with (pending on `storageType`).
//
// This means that the actual artifact is rarely available when this message
// is posted. But it is not unreasonable to assume that the artifact will
// will become available at some point later. Most signatures will expire in
// 30 minutes or so, forcing the uploader to call `createArtifact` with
// the same payload again in-order to continue uploading the artifact.
//
// However, in most cases (especially for small artifacts) it's very
// reasonable assume the artifact will be available within a few minutes.
// This property means that this exchange is mostly useful for tools
// monitoring task evaluation. One could also use it count number of
// artifacts per task, or _index_ artifacts though in most cases it'll be
// smarter to index artifacts after the task in question have completed
// successfully.
//
// See http://docs.taskcluster.net/queue/exchanges/#artifactCreated
type ArtifactCreated struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding ArtifactCreated) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding ArtifactCreated) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/artifact-created"
}

func (binding ArtifactCreated) NewPayloadObject() interface{} {
	return new(ArtifactCreatedMessage)
}

// When a task is successfully completed by a worker a message is posted
// this exchange.
// This message is routed using the `runId`, `workerGroup` and `workerId`
// that completed the task. But information about additional runs is also
// available from the task status structure.
//
// See http://docs.taskcluster.net/queue/exchanges/#taskCompleted
type TaskCompleted struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding TaskCompleted) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding TaskCompleted) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/task-completed"
}

func (binding TaskCompleted) NewPayloadObject() interface{} {
	return new(TaskCompletedMessage)
}

// When a task ran, but failed to complete successfully a message is posted
// to this exchange. This is same as worker ran task-specific code, but the
// task specific code exited non-zero.
//
// See http://docs.taskcluster.net/queue/exchanges/#taskFailed
type TaskFailed struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding TaskFailed) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding TaskFailed) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/task-failed"
}

func (binding TaskFailed) NewPayloadObject() interface{} {
	return new(TaskFailedMessage)
}

// Whenever TaskCluster fails to run a message is posted to this exchange.
// This happens if the task isn't completed before its `deadlìne`,
// all retries failed (i.e. workers stopped responding), the task was
// canceled by another entity, or the task carried a malformed payload.
//
// The specific _reason_ is evident from that task status structure, refer
// to the `reasonResolved` property for the last run.
//
// See http://docs.taskcluster.net/queue/exchanges/#taskException
type TaskException struct {
	RoutingKeyKind string `mwords:"*"`
	TaskId         string `mwords:"*"`
	RunId          string `mwords:"*"`
	WorkerGroup    string `mwords:"*"`
	WorkerId       string `mwords:"*"`
	ProvisionerId  string `mwords:"*"`
	WorkerType     string `mwords:"*"`
	SchedulerId    string `mwords:"*"`
	TaskGroupId    string `mwords:"*"`
	Reserved       string `mwords:"#"`
}

func (binding TaskException) RoutingKey() string {
	return generateRoutingKey(&binding)
}

func (binding TaskException) ExchangeName() string {
	return "exchange/taskcluster-queue/v1/task-exception"
}

func (binding TaskException) NewPayloadObject() interface{} {
	return new(TaskExceptionMessage)
}

func generateRoutingKey(x interface{}) string {
	val := reflect.ValueOf(x).Elem()
	p := make([]string, 0, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag
		if t := tag.Get("mwords"); t != "" {
			if v := valueField.Interface(); v == "" {
				p = append(p, t)
			} else {
				p = append(p, v.(string))
			}
		}
	}
	return strings.Join(p, ".")
}

type (
	// Message reporting a new artifact has been created for a given task.
	//
	// See http://schemas.taskcluster.net/queue/v1/artifact-created-message.json#
	ArtifactCreatedMessage struct {
		// Information about the artifact that was created
		Artifact struct {
			// Mimetype for the artifact that was created.
			ContentType string `json:"contentType"`
			// Date and time after which the artifact created will be automatically
			// deleted by the queue.
			Expires Time `json:"expires"`
			// Name of the artifact that was created, this is useful if you want to
			// attempt to fetch the artifact. But keep in mind that just because an
			// artifact is created doesn't mean that it's immediately available.
			Name string `json:"name"`
			// This is the `storageType` for the request that was used to create the
			// artifact.
			//
			// Possible values:
			//   * "s3"
			//   * "azure"
			//   * "reference"
			//   * "error"
			StorageType string `json:"storageType"`
		} `json:"artifact"`
		// Id of the run on which artifact was created.
		RunId  int                 `json:"runId"`
		Status TaskStatusStructure `json:"status"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
		// Identifier for the worker-group within which the run with the created
		// artifacted is running.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerGroup string `json:"workerGroup"`
		// Identifier for the worker within which the run with the created artifact
		// is running.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerId string `json:"workerId"`
	}

	// Message reporting that a task has complete successfully.
	//
	// See http://schemas.taskcluster.net/queue/v1/task-completed-message.json#
	TaskCompletedMessage struct {
		// Id of the run that completed the task
		RunId  int                 `json:"runId"`
		Status TaskStatusStructure `json:"status"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
		// Identifier for the worker-group within which this run ran.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerGroup string `json:"workerGroup"`
		// Identifier for the worker that executed this run.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerId string `json:"workerId"`
	}

	// Message reporting that a task has been defined. The task may or may not be
	// _scheduled_ too.
	//
	// See http://schemas.taskcluster.net/queue/v1/task-defined-message.json#
	TaskDefinedMessage struct {
		Status TaskStatusStructure `json:"status"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
	}

	// Message reporting that TaskCluster have failed to run a task.
	//
	// See http://schemas.taskcluster.net/queue/v1/task-exception-message.json#
	TaskExceptionMessage struct {
		// Id of the last run for the task, not provided if `deadline`
		// was exceeded before a run was started.
		RunId  int                 `json:"runId"`
		Status TaskStatusStructure `json:"status"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
		// Identifier for the worker-group within which the last attempt of the task
		// ran. Not provided, if `deadline` was exceeded before a run was started.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerGroup string `json:"workerGroup"`
		// Identifier for the last worker that failed to report, causing the task
		// to fail. Not provided, if `deadline` was exceeded before a run
		// was started.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerId string `json:"workerId"`
	}

	// Message reporting that a task failed to complete successfully.
	//
	// See http://schemas.taskcluster.net/queue/v1/task-failed-message.json#
	TaskFailedMessage struct {
		// Id of the run that failed.
		RunId  int                 `json:"runId"`
		Status TaskStatusStructure `json:"status"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
		// Identifier for the worker-group within which this run ran.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerGroup string `json:"workerGroup"`
		// Identifier for the worker that executed this run.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerId string `json:"workerId"`
	}

	// Message reporting that a task is now pending
	//
	// See http://schemas.taskcluster.net/queue/v1/task-pending-message.json#
	TaskPendingMessage struct {
		// Id of run that became pending, `run-id`s always starts from 0
		RunId  int                 `json:"runId"`
		Status TaskStatusStructure `json:"status"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
	}

	// Message reporting that a given run of a task have started
	//
	// See http://schemas.taskcluster.net/queue/v1/task-running-message.json#
	TaskRunningMessage struct {
		// Id of the run that just started, always starts from 0
		RunId  int                 `json:"runId"`
		Status TaskStatusStructure `json:"status"`
		// Time at which the run expires and is resolved as `failed`, if the run
		// isn't reclaimed.
		TakenUntil Time `json:"takenUntil"`
		// Message version
		//
		// Possible values:
		//   * 1
		Version int `json:"version"`
		// Identifier for the worker-group within which this run started.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerGroup string `json:"workerGroup"`
		// Identifier for the worker executing this run.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerId string `json:"workerId"`
	}

	// A representation of **task status** as known by the queue
	//
	// See http://schemas.taskcluster.net/queue/v1/task-status.json#
	TaskStatusStructure struct {
		// Deadline of the task, `pending` and `running` runs are resolved as **failed** if not resolved by other means before the deadline. Note, deadline cannot be more than5 days into the future
		Deadline Time `json:"deadline"`
		// Task expiration, time at which task definition and status is deleted. Notice that all artifacts for the must have an expiration that is no later than this.
		Expires Time `json:"expires"`
		// Unique identifier for the provisioner that this task must be scheduled on
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		ProvisionerId string `json:"provisionerId"`
		// Number of retries left for the task in case of infrastructure issues
		RetriesLeft int `json:"retriesLeft"`
		// List of runs, ordered so that index `i` has `runId == i`
		Runs []struct {
			// Reason for the creation of this run,
			// **more reasons may be added in the future**.
			//
			// Possible values:
			//   * "scheduled"
			//   * "retry"
			//   * "rerun"
			//   * "exception"
			ReasonCreated string `json:"reasonCreated"`
			// Reason that run was resolved, this is mainly
			// useful for runs resolved as `exception`.
			// Note, **more reasons may be added in the future**, also this
			// property is only available after the run is resolved.
			//
			// Possible values:
			//   * "completed"
			//   * "failed"
			//   * "deadline-exceeded"
			//   * "canceled"
			//   * "claim-expired"
			//   * "worker-shutdown"
			//   * "malformed-payload"
			//   * "resource-unavailable"
			//   * "internal-error"
			ReasonResolved string `json:"reasonResolved"`
			// Date-time at which this run was resolved, ie. when the run changed
			// state from `running` to either `completed`, `failed` or `exception`.
			// This property is only present after the run as been resolved.
			Resolved Time `json:"resolved"`
			// Id of this task run, `run-id`s always starts from `0`
			RunId int `json:"runId"`
			// Date-time at which this run was scheduled, ie. when the run was
			// created in state `pending`.
			Scheduled Time `json:"scheduled"`
			// Date-time at which this run was claimed, ie. when the run changed
			// state from `pending` to `running`. This property is only present
			// after the run has been claimed.
			Started Time `json:"started"`
			// State of this run
			//
			// Possible values:
			//   * "pending"
			//   * "running"
			//   * "completed"
			//   * "failed"
			//   * "exception"
			State string `json:"state"`
			// Time at which the run expires and is resolved as `failed`, if the
			// run isn't reclaimed. Note, only present after the run has been
			// claimed.
			TakenUntil Time `json:"takenUntil"`
			// Identifier for group that worker who executes this run is a part of,
			// this identifier is mainly used for efficient routing.
			// Note, this property is only present after the run is claimed.
			//
			// Syntax: ^([a-zA-Z0-9-_]*)$
			WorkerGroup string `json:"workerGroup"`
			// Identifier for worker evaluating this run within given
			// `workerGroup`. Note, this property is only available after the run
			// has been claimed.
			//
			// Syntax: ^([a-zA-Z0-9-_]*)$
			WorkerId string `json:"workerId"`
		} `json:"runs"`
		// Identifier for the scheduler that _defined_ this task.
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		SchedulerId string `json:"schedulerId"`
		// State of this task. This is just an auxiliary property derived from state
		// of latests run, or `unscheduled` if none.
		//
		// Possible values:
		//   * "unscheduled"
		//   * "pending"
		//   * "running"
		//   * "completed"
		//   * "failed"
		//   * "exception"
		State string `json:"state"`
		// Identifier for a group of tasks scheduled together with this task, by
		// scheduler identified by `schedulerId`. For tasks scheduled by the
		// task-graph scheduler, this is the `taskGraphId`.
		//
		// Syntax: ^[A-Za-z0-9_-]{8}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$
		TaskGroupId string `json:"taskGroupId"`
		// Unique task identifier, this is UUID encoded as
		// [URL-safe base64](http://tools.ietf.org/html/rfc4648#section-5) and
		// stripped of `=` padding.
		//
		// Syntax: ^[A-Za-z0-9_-]{8}[Q-T][A-Za-z0-9_-][CGKOSWaeimquy26-][A-Za-z0-9_-]{10}[AQgw]$
		TaskId string `json:"taskId"`
		// Identifier for worker type within the specified provisioner
		//
		// Syntax: ^([a-zA-Z0-9-_]*)$
		WorkerType string `json:"workerType"`
	}
)

// Wraps time.Time in order that json serialisation/deserialisation can be adapted.
// Marshaling time.Time types results in RFC3339 dates with nanosecond precision
// in the user's timezone. In order that the json date representation is consistent
// between what we send in json payloads, and what taskcluster services return,
// we wrap time.Time into type queueevents.Time which marshals instead
// to the same format used by the TaskCluster services; UTC based, with millisecond
// precision, using 'Z' timezone, e.g. 2015-10-27T20:36:19.255Z.
type Time time.Time

// MarshalJSON implements the json.Marshaler interface.
// The time is a quoted string in RFC 3339 format, with sub-second precision added if present.
func (t Time) MarshalJSON() ([]byte, error) {
	if y := time.Time(t).Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("queue.Time.MarshalJSON: year outside of range [0,9999]")
	}
	return []byte(`"` + t.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// The time is expected to be a quoted string in RFC 3339 format.
func (t *Time) UnmarshalJSON(data []byte) (err error) {
	// Fractional seconds are handled implicitly by Parse.
	x := new(time.Time)
	*x, err = time.Parse(`"`+time.RFC3339+`"`, string(data))
	*t = Time(*x)
	return
}

// Returns the Time in canonical RFC3339 representation, e.g.
// 2015-10-27T20:36:19.255Z
func (t Time) String() string {
	return time.Time(t).UTC().Format("2006-01-02T15:04:05.000Z")
}
