# lms-notifier

A simple tool to keep track of lms modules

## Getting Started

These instructions will get you a copy of the project up and running on your local machine.

### Prerequisites

You must have Golang and Git installed on your machine.

Golang: https://go.dev/doc/install
Git:https://git-scm.com/downloads

### Cloning

```
git clone https://github.com/ArashSameni/lms-notifier
cd lms-notifier
```

## Running

```
go mod tidy
go run main.go -u "Username" -p "Password"
```
The first run of application will create multiple files in the sources folder, further runnings will compare and extract changes from your lms account.

## Contributing

Pull requests are welcome.
