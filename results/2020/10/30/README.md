# dispatch_latency
Message latency studies for Qpid Dispatch Router


## Date 
  
  2020 10 30

## Versions
  * dispatch: 1.15.0-SNAPSHOT  
  * proton: 0.33.0-SNAPSHOT  


### This setup is meant to simulate a potential user's architecture.

  * There is 1 router.
  * Each client is a separate running process (the program written
    in C to the Proactor interface.)
  * Each client has 1 connection with 1 link to the router.
  * Each receiver gets its own address.
  * There are 40 senders for each receiver.
  * Each sender sends 10 messages per second.
  * Each message has 100 bytes of payload.
  * The messages are not presettled.
  * The router has 16 worker threads.
  * The messages are *not* pre-settled.
  * The tests were set up and run by the Mercury testing system. ( https://github.com/mgoulish/mercury ).

 <br/>

### Hardware

  * AMD Ryzen Threadripper 3970X 32-Core 3.7 GHz
  * 256 GB RAM
  
<br/>

### The Tests

  * 1 router, N senders, N receivers, N addresses -- all colocated on one box, where N = { 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 9500, 10000 }
  * Each message has a timestamp placed into the body by the sender just before send-time.
  * The receiver takes a time reading immediately after receiving each message, then parses out the send-timestamp, finds the flight time, and stores that in a large pre-allocated array.
  * No memory allocation by the client code at runtime.
  * Also no printing by the clients at runtime. Flight times are stored in an array and printed out only after the test is complete.


<br/>

### The Graphs

  * The results are in two graphs. The first shows results from testing 1000 sender-receiver pairs, up to 9000.  The second contains all the same data as the first, but it also includes the last two tests: 9500 and 10000.   ( The last two tests change the scale quite a bit. )

  * Both graphs show two curves: the mean latency for all messages, and the 99% line -- the timing that includes 99% of all messages.

  * The graphs are in this directory: 2020_10_30_latency.jpg and 2020_10_30_latency_first_nine.jpg


### Note on speed
If these timings seem faster than what you remember, it's because they are. In the past I mistakenly allowed senders to start sending even while other clients were still starting up and attaching to the router. That caused some very slow timings near the beginning of the test, which distorted results.









