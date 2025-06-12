# Blazium Services

Blazium Services are a suite of services designed to simplify game development
with aspects such as multiplayer, authentication, and more. Whether you’re a solo
developer working on a casual multiplayer game or a studio planning an ambitious
MMO, Blazium offers scalable, lightweight services to meet your needs.
They are designed to integrate inside the Blazium engine and work on
all platforms (including desktop, mobile and web).

When developing our first game, Project Hangman, we found a lot of shortcomings
in the services that exist today for making games. Our needs were simple, or so
we thought: Make a multiplayer hangman game that scales up to 10.000 players or
more. (and in the future, make a MMO that scales to millions)

Blazium Services were born out of our frustration
with existing platforms. Our goal was to create a lightweight, scalable, and
cross-platform solution that meets modern developers’ needs—without the
unnecessary complexity or high costs.

    +---------------------+               +---------------------+
    |   Blazium Engine    |   HTTP and    |   Blazium Service   |
    |---------------------|   WebSocket   |---------------------|
    | - Send requests     |  <--------->  | - Handle requests   |
    | - Receive responses |               | - Send responses    |
    +---------------------+               +---------------------+

As such we thought of creating our first service, the Lobby Service, whose
purpose is to handle matchmaking between users, while not giving away the IP.
The system is designed to scale to a large number of users, while using low
resources. From there we went ahead and built the Scripted Lobby Service, the
Login Service, the Master Server Service, Leaderboard Service, and with
increasing needs we added to our roadmap more services.

### Features
- **Fast**: Our services are designed to be as small as possible, working
well whether you design a small multiplayer game or a large-scale MMO.
Cross-platform. Supports desktop, mobile, web and consoles, ensuring your game
can reach players everywhere.
- **Lightweight**: Designed to scale to thousands with
very few resources, while being lightweight.
- **Self-hosting**: Need full control?
We also offer the option to self host and easy configuration of the services you
use, if the free hosted services are not enough.

Blazium engine comes out-of-the-box with these services as nodes,
so that integration is easy and quick, with no manual setup needed.
We also have a free server on our domain that the nodes connect to automatically,
so that you can start using our services as soon as you need it.

### Services
Here is the list of currently available services with a link to their documentation.
- [Lobby](https://docs.blazium.app/classes/class_lobbyclient.html)
- [Scripted Lobby](https://docs.blazium.app/classes/class_authoritativelobbyclient.html)
- [Login](https://docs.blazium.app/classes/class_loginclient.html)
- [Master Server](https://docs.blazium.app/classes/class_masterserverclient.html)