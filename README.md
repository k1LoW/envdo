# envdo

Execute commands with environment variables from .env files.

## Features

- Load environment variables from `.env` files
- Support profile-based environment files (`.env.{profile}`)
- Priority-based directory search (current directory > `$XDG_CONFIG_HOME/envdo`)
- Execute any command with loaded environment variables
- Display loaded environment variables in export format

## Install

### go install

```console
$ go install github.com/k1LoW/envdo@latest
```

### Binary releases

Download from [releases page](https://github.com/k1LoW/envdo/releases).

## Usage

### Basic usage

```console
$ envdo -- echo $MY_VAR
```

### With profile

```console
$ envdo --profile production -- node app.js
$ envdo -p dev -- npm start
```

### Show loaded environment variables

```console
$ envdo
export API_KEY=your_api_key
export DATABASE_URL=postgresql://localhost/mydb
```

## .env files

envdo searches for `.env` files in the following directories in order of priority:

1. Current directory
2. `$XDG_CONFIG_HOME/envdo` (typically `~/.config/envdo`)

### Basic .env file format

```
# Comments are supported
API_KEY=your_api_key
DATABASE_URL=postgresql://localhost/mydb

# Quoted values are supported
SECRET="my secret value"
ANOTHER_SECRET='another secret'
```

### Profile-based .env files

When using the `--profile` option, envdo looks for `.env.{profile}` files:

```console
$ envdo --profile production -- node app.js
# Loads .env.production
```
