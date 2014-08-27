Go Chat
=======

Simple TCP chat program written in Golang. Demonstrate how easy network communication is with Go. This is a precursor to a more complicated distributed computing system I am writing. 

The program sends messages to anybody who has connected to you and anybody you have connected to.

Listens on port 1337 for incoming connections.

Two GoChat programs can communicate with eachother, or you can connect to a GoChat program using telnet. If you just want to test it locally, you can always launch GoChat and type 'telnet 127.0.0.1 1337' from terminal.


Commands
--------

Type 'connect [ip]:[port]' from within the program to connect to another computer.

Type 'quit' to exit.