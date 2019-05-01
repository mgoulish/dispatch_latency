# dispatch_latency
Message latency studies for Qpid Dispatch Router


## 1.6.0 Late April, 2019

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

  * 16 cores on a single chip. 32 threads.
  * 3.0 GHz
  * AMD Ryzen Threadripper 1950X
  * 64 GB RAM
  
<br/>

### The Tests

  * 7 tests, varying the number of receivers: 30, 50, 70, 100, 150, 170, 200.
  * Each test has simillar number of total messages sent, always close to 10 million.
  * Each message has a timestamp placed into the body by the sender just before send-time.
  * The receiver takes a time reading immediately after receiving each message, then parses out the send-timestamp, finds the flight time, and stores that in a large pre-allocated array.
  * No memory allocation by the client code at runtime.
  * Also no printing by the clients at runtime. Flight times are stored in a large array and printed out only after the test is complete.


<br/>

### The Graphs

  * Data is processed into graphable form by some Go code ( Checked in here, but soon to become part of the Mercury testing system. )
  * And then turned into graphs by gnuplot.








