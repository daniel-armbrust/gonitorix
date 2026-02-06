![Gonitorix Logo](docs/img/gonitorix-logo.png)


Gonitorix
=========

Gonitorix is a lightweight system and network monitoring tool written in Go.

It periodically collects system and network statistics, stores historical data in RRD databases, and generates graphs for visualization — inspired by the Monitorix monitoring system.

Gonitorix was developed as an **independent implementation** and does not reuse code from Monitorix.

Features
--------

- System load, memory, processes, entropy and uptime monitoring
- Network interface statistics
- RRD-based historical storage
- Automatic graph generation
- YAML configuration file
- Auto-discovery of network interfaces
- Modular design
- Written in Go
- GPLv3 licensed

Inspiration
-----------

This project is inspired by the Monitorix monitoring tool:

https://www.monitorix.org/

Monitorix is licensed under the GNU General Public License version 2.

Gonitorix is **not affiliated with or endorsed by the Monitorix project** and was implemented independently.

License
-------

Gonitorix is licensed under the GNU General Public License version 3 (GPLv3).

See the `LICENSE` file for full license text.

Credits
-------

Gonitorix was created by [Daniel Armbrust](https://www.linkedin.com/in/daniel-armbrust/).
You can contact me at [darmbrust@gmail.com](mailto:darmbrust@gmail.com).