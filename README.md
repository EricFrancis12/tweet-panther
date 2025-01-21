# Tweet Panther

A configurable Twitter API microservice.

## Environment Variables

Run the following command to create a new `.env` file at the root of the project.

```bash
make create_env
```

### Required Environment Variables

- `PORT`
- `AUTH_TOKEN`
- `USER_ID`
- `API_KEY`
- `API_KEY_SECRET`
- `O_AUTH_TOKEN`
- `O_AUTH_TOKEN_SECRET`

See `.env.example` for detailed explanations of each varaible.

__Note:__ An `.env.local` file at the project root will override any variables present in the `.env` file.

## Usage

Build and run the application:

```bash
make run
```

Build the executable without running:

```bash
make build
```

## License

[MIT](https://mit-license.org)
