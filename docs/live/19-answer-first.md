# Answering before the write ends (T90's probe)

**Measured: Cloud Run hands a tool's answer to the client as soon as it is written, keeps the request open and the instance's CPU whole until the handler returns, even once the client has gone, and Claude goes on with the answer without waiting for the request to end.** 2026-10-07. A probe of about a hundred lines on go-sdk v1.8.0, deployed for the run as a service of its own in `mathtrail-prod` (`us-central1`, CPU allocated only while a request is served, one instance), and deleted with its image afterwards.

- **The answer arrives at once.** Over HTTP/1.1 and over HTTP/2 alike, the event with the result came 0.5 s after the request was sent, the whole round trip, and the response ended 15.1 s later, when the probe let the request go.
- **The CPU stays while the request is held.** A fixed loop of hashing took 8–12 ms every 250 ms (median 9), as it did at the start, and an HTTPS call to Google took 8–34 ms.
- **A client that leaves goes unseen.** With the client gone 1 s after the answer, the probe's request still stayed open for the full 15 s: the platform did not cancel it, and the loop and the calls kept their times (8–11 ms, 8–10 ms). The container cannot tell whether its client is still there.
- **Claude does not wait for the end.** Asked in claude.ai to call the probe twice in a row, its client sent the second call 2.96 s after the first answer, while the first request still had 12 s to go. With the two requests held at once on one instance, the loop took 8–24 ms and the calls 7–15 ms.
- **Not measured here:** ChatGPT, whose client may behave otherwise.

## What this is

T90 ([RUN.md](../../RUN.md)) sets the service's own share of a task's path under a second. Since T88 and T92 a task written ahead is handed out by `next_task` on the card it draws. Of that call's 2.1 s, T85 measured about 1.4 s as the profile's write to Drive ([16-task-time](16-task-time.md)). Recording an answer is built the same way. The child need not wait for those writes, if the service answers first and finishes the write within the same request.

Whether that helps rests on three things no documentation settles:
- whether Cloud Run's front end passes an event of a stream at once, or holds it until the response ends;
- whether the instance keeps its CPU while a request is held open after its answer, and after its client has gone;
- whether a chat host acts on a tool's result as it arrives, or waits for the request to end.

## The probe

- **The server.** go-sdk v1.8.0 serves stateless Streamable HTTP answering with a stream of events, as the service does. It has one tool, `probe`, which answers with the server's time.
- **The hold.** A wrapper around the protocol keeps a `tools/call` open for 15 s after the SDK has written its answer. Every 250 ms it times a fixed loop: 200 rounds of SHA-256 over 64 KiB. At +1 s and +5 s it times an HTTPS call to `www.google.com/generate_204` on a fresh connection. It logs when the answer was written and when the request's context ended, and keeps its lines in memory to be read back, so the project's log was not read.
- **Where it ran.** It was built from the images the repository pins and pushed to the project's registry. It was deployed with `--cpu-throttling --max-instances 1 --allow-unauthenticated`. After the runs the service and its image were deleted. Its code is kept outside the repository, as T85's tools are.

## The runs

| Run | Client | The answer arrived | The request ended | Loop, median / max | HTTPS calls |
|---|---|---:|---:|---|---|
| 1 | curl, HTTP/1.1 | 509 ms | 15,641 ms | 9 / 12 ms | 34, 10 ms |
| 1 | curl, HTTP/2 | 524 ms | 15,640 ms | 9 / 12 ms | 8, 8 ms |
| 2 | curl, gone 1 s after the answer | at once | 15,090 ms after the answer; the client's leaving never reached the probe | 8 / 11 ms | 8, 10 ms |
| 3 | claude.ai, two calls in a row | the second call came 2.96 s after the first answer | 15.1 s after each answer | 9 / 24 ms | 13, 15, 9, 7 ms |

- Run 1's times are counted from the request's start, and the other columns from the answer.
- Both protocols reached the container as HTTP/1.1.
- In run 3 the connector's setup came from `python-httpx/0.28.1` and the calls from `Claude-User`. The model said the two server times were about three seconds apart, and it noticed that the hold had not delayed the second call.

## What it means for T90

- **Answering first saves the write's time in Claude.** The card shows the task as the answer arrives, and the model goes on while the write finishes.
- **The write survives a client that leaves.** The request stays open and the CPU whole until the handler returns, so a write finished after the answer is not starved.
- **Whether the client stayed cannot be logged.** The container is not told, so a field for it would always say it stayed.
- **The next call of the same account can come while the write still runs.** In run 3 it came 2.96 s later, on the same instance. So a call has to wait for the writes of its account that the instance has begun, or it reads the file as it was.
