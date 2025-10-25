package output

import "golang.org/x/text/language"

type translations map[string]string

var en = translations{
	"instance":    "Manage Minecraft instances",
	"auth":        "Manage account authentication",
	"about":       "Display launcher version and about",
	"list":        "List all instances",
	"completions": "Outputs shell command to install completions",

	"login":               "Login in to an account",
	"login.code.fetching": "Loading device code...",
	"login.code":          "Use the code %s at %s to sign in",
	"login.browser":       "A web browser will be opened to continue authenticatication.",
	"login.url":           "If the browser does not open, please copy and paste this URL into your browser: %s",
	"login.complete":      "Logged in as %s",
	"login.redirect":      "Logged in! You can close this window and return to the launcher.",
	"login.redirectfail":  "Failed to log in: An error occurred during authentication.",
	"login.arg.nobrowser": "Use device code instead of browser for authentication",

	"logout":          "Log out of an account",
	"logout.complete": "Logged out from account.",

	"create":                   "Create a new instance",
	"create.complete":          "Created instance '%s' with Minecraft %s (%s%s)",
	"create.arg.id":            "Instance name",
	"create.arg.loader":        "Mod loader",
	"create.arg.version":       "Game version",
	"create.arg.loaderversion": "Mod loader version",

	"delete":          "Delete an instance",
	"delete.confirm":  "Are you sure you want to delete this instance?",
	"delete.warning":  "'%s' will be lost forever (a long time!) [y/n] ",
	"delete.complete": "Deleted instance '%s'",
	"delete.abort":    "Operation aborted",
	"delete.arg.id":   "Instance to delete",
	"delete.arg.yes":  "Assume yes to all questions",

	"rename":          "Rename an instance",
	"rename.complete": "Renamed instance.",
	"rename.arg.id":   "Instance to rename",
	"rename.arg.new":  "New name for instance",

	"search":               "Search versions",
	"search.complete":      "Found %d entries",
	"search.table.version": "Version",
	"search.table.type":    "Type",
	"search.table.date":    "Release Date",
	"search.table.name":    "Name",
	"search.arg.query":     "Search query",
	"search.arg.kind":      "What to search for",
	"search.arg.reverse":   "Reverse the listing",

	"start":                    "Start the specified instance",
	"start.arg.id":             "Instance to launch",
	"start.arg.verbose":        "Increase verbosity",
	"start.arg.username":       "Set username (offline mode)",
	"start.arg.server":         "Join a server upon starting the game",
	"start.arg.world":          "Join a world upon starting the game",
	"start.arg.demo":           "Start the game in demo mode",
	"start.arg.disablemp":      "Disable multiplayer",
	"start.arg.disablechat":    "Disable chat",
	"start.arg.width":          "Game window width",
	"start.arg.height":         "Game window height",
	"start.arg.jvm":            "Path to the JVM",
	"start.arg.jvmargs":        "Extra JVM arguments",
	"start.arg.minmemory":      "Minimum memory",
	"start.arg.maxmemory":      "Maximum memory",
	"start.arg.prepare":        "Install all necessary resources but do not start the game.",
	"start.arg.opts":           "Game Options",
	"start.arg.overrides":      "Configuration Overrides",
	"start.prepared":           "Game prepared successfully.",
	"start.processing":         "Post processors are being run. This may take some time.",
	"start.launch.downloading": "Downloading files",
	"start.launch.assets":      "Identified %d assets",
	"start.launch.libraries":   "Identified %d libraries",
	"start.launch.metadata":    "Version metadata retrieved",
	"start.launch.jvmargs":     "JVM arguments: %s",
	"start.launch.gameargs":    "Game arguments: %s",
	"start.launch.info":        "Starting main class %q. Game directory is %q.",
	"start.launch":             "Launching game as %s",

	"arg.verbosity": "Increase launcher output verbosity",
	"arg.dir":       "Root directory for launcher files",
	"arg.nocolor":   "Disable all color output. The NO_COLOR environment variable is also supported.",

	"tip.internet":  "Check your internet connection.",
	"tip.cache":     "Remote resources were not cached and were unable to be retrieved. Check your Internet connection.",
	"tip.configure": "Configure this instance with the `instance.toml` file within the instance directory.",
	"tip.nojvm":     "If a Mojang-provided JVM is not available, you can install it yourself and set the path to the Java executable in the instance configuration.",
	"tip.noaccount": "To launch in offline mode, use the --username (-u) flag.",

	"launcher.description": "A minimal command-line Minecraft launcher.",
	"launcher.license":     "Licensed MIT",
	"launcher.copyright":   "Copyright 2024-2025 telecter",
	"launcher.error":       "Error",
	"launcher.warning":     "Warning",
	"launcher.debug":       "Debug",
	"launcher.tip":         "Tip",
}

var de = translations{
	"instance":    "Minecraft-Instanze verwalten",
	"auth":        "Konto-Authentifizierung verwalten",
	"about":       "Version und andere Informationen anzeigen",
	"list":        "Alle Instanze auflisten",
	"completions": "Befehl ausstoßen, der Tab-Vervollständigungen einrichtet",

	"login":               "Anmelden",
	"login.code.fetching": "Gerätcode laden ...",
	"login.code":          "Verwende den Code %s auf %s, um dich anzumelden.",
	"login.browser":       "Ein Webbrowser wird geöffnet, um Authentifizierung fortzufahren.",
	"login.url":           "Falls der Webbrowser nicht öffnet, öffne diesen URL: %s",
	"login.complete":      "Angemeldet als %s",
	"login.redirect":      "Angemeldet! Du kannst dieses Fenster schließen und zum Launcher zurückkehren.",
	"login.redirectfail":  "Anmeldung fehlgeschlagen: Ein Fehler ist während der Authentifizierung aufgetreten.",
	"login.arg.nobrowser": "Gerätcode statt Webbrowser für Authentifizierung verwenden.",

	"logout":          "Abmelden",
	"logout.complete": "Abgemeldet.",

	"create":                   "Neue Instanze erstellen",
	"create.complete":          "Instanz '%s' mit Minecraft %s (%s%s) erstellt",
	"create.arg.id":            "Instanzname",
	"create.arg.loader":        "Mod Loader",
	"create.arg.version":       "Spielversion",
	"create.arg.loaderversion": "Mod Loader Version",

	"delete":          "Instanze löschen",
	"delete.confirm":  "Bist du sicher, dass du diese Instanz löschen willst?",
	"delete.warning":  "'%s' wird für immer verloren sein (eine lange Zeit!) [y/n] ",
	"delete.complete": "Instanz '%s' gelöscht",
	"delete.abort":    "Abgebrochen.",
	"delete.arg.id":   "Instanz zum Löschen",
	"delete.arg.yes":  "Zu allen Fragen automatisch zustimmen.",

	"rename":          "Instanze umbenennen",
	"rename.complete": "Instanz umbennant.",
	"rename.arg.id":   "Instanz zum Umbenennen",
	"rename.arg.new":  "Neuen Name für die Instanz",

	"search":               "Versionen suchen",
	"search.complete":      "%d Ergebnise gefunden",
	"search.table.version": "Version",
	"search.table.type":    "Typ",
	"search.table.date":    "Veröffentlicht am",
	"search.table.name":    "Name",
	"search.arg.query":     "Suchanfrage",
	"search.arg.kind":      "Suchtyp",
	"search.arg.reverse":   "Liste umgekehrt anzeigen",

	"start":                    "Instanze starten",
	"start.arg.id":             "Instanz zum Starten",
	"start.arg.username":       "Benutzername (Offlinemodus)",
	"start.arg.server":         "Einem Server beim Spielstart beitreten",
	"start.arg.world":          "Einer Welt beim Spielstart beitreten ",
	"start.arg.demo":           "Spiel im Testmodus starten",
	"start.arg.disablemp":      "Mehrspielermodus deaktivieren",
	"start.arg.disablechat":    "Chat deaktivieren",
	"start.arg.width":          "Spielfensterbreite",
	"start.arg.height":         "Spielfensterhöhe",
	"start.arg.jvm":            "JVM-Pfad",
	"start.arg.jvmargs":        "JVM Argumente",
	"start.arg.minmemory":      "Minimale Arbeitsspeicherauslastung",
	"start.arg.maxmemory":      "Maximale Arbeitsspeicherauslastung",
	"start.arg.prepare":        "Alle gebrauchten Spielressourcen herunterladen, aber das Spiel nicht starten.",
	"start.arg.opts":           "Spieleinstellungen",
	"start.arg.overrides":      "Konfigurationüberschreibungen",
	"start.prepared":           "Spiel erfolgreich vorbereitet.",
	"start.processing":         "Nachbearbeitungen sind jetzt im Gange. Das kann einige Zeit dauern.",
	"start.launch.downloading": "Dateien herunterladen ...",
	"start.launch.assets":      "%d Ressourcen identifiziert",
	"start.launch.libraries":   "%d Bibliotheken identifiziert",
	"start.launch.metadata":    "Versiondaten heruntergeladen",
	"start.launch.jvmargs":     "JVM Argumente: %s",
	"start.launch.gameargs":    "Spielargumente: %s",
	"start.launch.info":        "Hauptklasse %q wird gestartet. Spielverzeichnis ist %q.",
	"start.launch":             "Spiel als %s starten ...",

	"arg.verbosity": "Gesprächigkeit ändern",
	"arg.dir":       "Wurzelverzeichnis für Launcherdateien",
	"arg.nocolor":   "Farben nicht anzeigen. Die NO_COLOR Umgebungsvariable kann auch benutzt werden.",

	"tip.internet":  "Stell sicher, dass deine Internetverbindung funktioniert.",
	"tip.cache":     "Onlineressourcen waren nicht im Cache und konnten nicht heruntergeladen werden. Überprüfe deine Internetverbindung.",
	"tip.configure": "Die Einstellungen dieser Instanz können in der `instance.toml` Datei im Instanzverzeichnis angepasst werden.",
	"tip.nojvm":     "Falls ein JVM von Mojang nicht verfügbar ist, kannst du es selbst installieren und den Pfad zur Java Datei in der Instanzkonfiguration einstellen.",
	"tip.noaccount": "Um in Offlinemodus zu starten, verwende den --username (-u) Parameter.",

	"launcher.description": "Ein minimalisticher Minecraft Launcher für die Command Line.",
	"launcher.license":     "MIT-Lizenz",
	"launcher.copyright":   "Copyright 2024-2025 telecter",
	"launcher.error":       "Fehler",
	"launcher.warning":     "Warnung",
	"launcher.debug":       "Debug",
	"launcher.tip":         "Tip",
}

var lang = en

// SetLang changes the language to the specified language, if translations for it exist.
func SetLang(tag language.Tag) {
	switch tag {
	case language.German:
		lang = de
	default:
		lang = en
	}
}

// Translations returns the map of output translations for the current language.
func Translations() map[string]string {
	return lang
}

// Translate takes a translation string and looks up its human-readable text. If not available, it returns the same translation string.
func Translate(key string) string {
	t, ok := lang[key]
	if !ok {
		return key
	}
	return t
}
