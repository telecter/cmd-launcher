# cmd-launcher API (1.1.0)
In addition to the CLI, this launcher also provides an API which can be used to programmatically interact with it.
This page will be structured similarly to the main README, however for the API not the CLI.

### Creating an instance
You will first need to create an instance to start the game. For example:

```go
inst, err := launcher.CreateInstance(launcher.InstanceOptions{
		GameVersion: "1.21.5",
		Name:        "MyInstance",
		Loader:      launcher.LoaderQuilt,
		LoaderVersion: "latest",
})
```
The `Loader` can be `LoaderQuilt`, `LoaderFabric`, or `LoaderVanilla`.

`LoaderVersion`, if the mod loader is not vanilla, can either be set to `latest` for the latest version or a version of that specific mod loader.

### Getting an instance
If you have an instance that is already created, you can use the `FetchInstance` function to get it.
```go
inst, err := launcher.FetchInstance("MyInstance")
```
Instances can be removed with the `launcher.RemoveInstance` function which follows the same structure.

They can also be renamed with their `.Rename` method.
### Preparing the game

After you create your instance, in order to start the game, you will need to prepare an launch environment.

```go
env, err := launcher.Prepare(inst, launcher.EnvOptions{
	Session: auth.Session{
		Username: "Dinnerbone",
	},
	Config: inst.Config,
}, myWatcher{})
```

The `Config` value can also be changed before it is used. In this case, the `inst.Config` from the instance is used. However, you could change a value if you want,
```go
config := inst.Config
config.WindowResolution.Height = 200
...
```

Now what about the `myWatcher{}`? The watcher is used to respond to various events that occur during the instance preparation. You need to define a struct on the EventWatcher interface.

```go
type myWatcher struct {}

func (myWatcher) Handle(event any) {
	switch e := event.(type) {
	case launcher.DownloadingEvent:
		...
    case launcher.LibrariesResolvedEvent:
        ...
    case launcher.AssetsResolvedEvent:
        ...
	}
}
```
This could be used, for example, to create a progress bar of the libraries/assets download progress.

The `Session` field is set without an access token, meaning it's an offline session. We will get to authenticating later.

### Starting the game
After you have a launch environment from `launcher.Prepare` you will need to create a `Runner` to actually run and monitor the game in the way that you want.

```go
type myRunner struct{}

func (runner) Run(cmd *exec.Cmd) error {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
```
This runner takes in an `*exec.Cmd` and runs it, copying its I/O streams to the console. This use case would likely be quite common, so there is a `launcher.ConsoleRunner` implementation which does exactly this.  
However, you should use your own runner if you have a different use case, such as copying logs to JSON, etc...

With this runner, you can start the game.
```go
err := launcher.Launch(env, myRunner{})
```

### Authentication
In order to authenticate, you will need to have a Microsoft Azure app. After creating that, copy the Client ID for use here. You will likely also want to select a localhost redirect URI in the Azure dashboard. **Make sure to include a port to use!**

Initialize the values like so:
```go
// You will want to put this in the init() function or sometime before you run any authentication functions
func init() {
	auth.MSA.ClientID = "your client ID"
	// needed if you want to use the auth code flow, which requires a redirect
	auth.MSA.RedirectURI = "your redirect URI"
}
```



**OAuth2 auth code flow**  
First, you need to get an auth URL for the user to go to.
```go
url := auth.MSA.AuthCodeURL()
```
Then you will either need to display this link to the user, or open a browser window for the user to authenticate. After this, wait for the user to authenticate with the redirect:
```go
session, err := auth.AuthenticateWithRedirect()
```
And then you'll have your session!  

**OAuth2 device code flow**   
First, get a device code and auth link for the user to navigate to.
```go
resp, err := auth.MSA.FetchDeviceCode()
```
A `deviceCodeResponse` is returned. The `UserCode` and `VerificationURI` should be displayed to the user so that they can authenticate.  
Then you should poll the endpoint for authentication updates to get your session:

```go
session, err := auth.AuthenticateWithCode(resp)
```

**Refreshing**   
You definitely would not want to use a link or auth code every time, so if you already have an MSA refresh token, you can simply refresh the authentication data. To do this, use the `Authenticate` function:

```go
session, err := auth.Authenticate()
```
This will refresh any expired tokens and data and give you a session.  
And that's it! You can use this session in the `launcher.Prepare` function. The authentication data is automatically saved to the env.AuthStorePath file.  