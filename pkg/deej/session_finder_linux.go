package deej

import (
	"fmt"
	"net"

	"github.com/jfreymuth/pulse/proto"
	"go.uber.org/zap"
)

type paSessionFinder struct {
	logger        *zap.SugaredLogger
	sessionLogger *zap.SugaredLogger

	client *proto.Client
	conn   net.Conn
}

func newSessionFinder(logger *zap.SugaredLogger) (SessionFinder, error) {
	client, conn, err := proto.Connect("")
	if err != nil {
		logger.Warnw("Failed to establish PulseAudio connection", "error", err)
		return nil, fmt.Errorf("establish PulseAudio connection: %w", err)
	}

	request := proto.SetClientName{
		Props: proto.PropList{
			"application.name": proto.PropListString("deej"),
		},
	}
	reply := proto.SetClientNameReply{}

	if err := client.Request(&request, &reply); err != nil {
		return nil, err
	}

	sf := &paSessionFinder{
		logger:        logger.Named("session_finder"),
		sessionLogger: logger.Named("sessions"),
		client:        client,
		conn:          conn,
	}

	sf.logger.Debug("Created PA session finder instance")

	return sf, nil
}

func (sf *paSessionFinder) GetAllSessions() ([]Session, error) {
	sessions := []Session{}

	// get the master sink session
	masterSink, err := sf.getMasterSinkSession()
	if err == nil {
		sessions = append(sessions, masterSink)
	} else {
		sf.logger.Warnw("Failed to get master audio sink session", "error", err)
	}

	// get the master source session
	masterSource, err := sf.getMasterSourceSession()
	if err == nil {
		sessions = append(sessions, masterSource)
	} else {
		sf.logger.Warnw("Failed to get master audio source session", "error", err)
	}

	// enumerate sink inputs and add sessions along the way
	if err := sf.enumerateAndAddSessions(&sessions); err != nil {
		sf.logger.Warnw("Failed to enumerate audio sessions", "error", err)
		return nil, fmt.Errorf("enumerate audio sessions: %w", err)
	}

	return sessions, nil
}

func (sf *paSessionFinder) Release() error {
	if err := sf.conn.Close(); err != nil {
		sf.logger.Warnw("Failed to close PulseAudio connection", "error", err)
		return fmt.Errorf("close PulseAudio connection: %w", err)
	}

	sf.logger.Debug("Released PA session finder instance")

	return nil
}

func (sf *paSessionFinder) getMasterSinkSession() (Session, error) {
	request := proto.GetSinkInfo{
		SinkIndex: proto.Undefined,
	}
	reply := proto.GetSinkInfoReply{}

	if err := sf.client.Request(&request, &reply); err != nil {
		sf.logger.Warnw("Failed to get master sink info", "error", err)
		return nil, fmt.Errorf("get master sink info: %w", err)
	}

	// create the master sink session
	sink := newMasterSession(sf.sessionLogger, sf.client, reply.SinkIndex, reply.Channels, true)

	return sink, nil
}

func (sf *paSessionFinder) getMasterSourceSession() (Session, error) {
	request := proto.GetSourceInfo{
		SourceIndex: proto.Undefined,
	}
	reply := proto.GetSourceInfoReply{}

	if err := sf.client.Request(&request, &reply); err != nil {
		sf.logger.Warnw("Failed to get master source info", "error", err)
		return nil, fmt.Errorf("get master source info: %w", err)
	}

	// create the master source session
	source := newMasterSession(sf.sessionLogger, sf.client, reply.SourceIndex, reply.Channels, false)

	return source, nil
}

func (sf *paSessionFinder) enumerateAndAddSessions(sessions *[]Session) error {
	request := proto.GetSinkInputInfoList{}
	reply := proto.GetSinkInputInfoListReply{}

	if err := sf.client.Request(&request, &reply); err != nil {
		sf.logger.Warnw("Failed to get sink input list", "error", err)
		return fmt.Errorf("get sink input list: %w", err)
	}

	for _, info := range reply {
		fmt.Fprintf(os.Stderr, "[DEBUG] Discovered PulseAudio stream index %d\n", info.SinkInputIndex)
		for k, v := range info.Properties {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", k, v)
		}

		name := sf.getBestSessionName(info.Properties)
		fmt.Fprintf(os.Stderr, "  - CHOSEN NAME: %s\n", name)

		if name == "" {
			sf.logger.Warnw("Failed to get sink input's process name",
				"sinkInputIndex", info.SinkInputIndex)

			continue
		}

		// Log at info level for now to ensure we see it
		sf.logger.Infow("Found audio session",
			"index", info.SinkInputIndex,
			"chosenName", name,
			"propertyKeys", info.Properties.Keys())

		// create the deej session object
		newSession := newPASession(sf.sessionLogger, sf.client, info.SinkInputIndex, info.Channels, name)

		// add it to our slice
		*sessions = append(*sessions, newSession)

	}

	return nil
}

func (sf *paSessionFinder) getBestSessionName(props proto.PropList) string {

	// 1. try the flatpak app ID if it exists (e.g. "org.mozilla.firefox")
	// we want the last part of it
	appID, ok := props["pipewire.access.portal.app_id"]
	if ok {
		parts := strings.Split(appID.String(), ".")
		return parts[len(parts)-1]
	}

	// list of generic process names that we should try to resolve to something better
	genericNames := []string{"bwrap", "xdg-desktop-portal", "flatpak"}

	// 2. try the actual binary name
	binaryName, ok := props["application.process.binary"]
	if ok {
		name := binaryName.String()

		// if it's not generic and doesn't end in -bin, we're done
		isGeneric := false
		for _, generic := range genericNames {
			if name == generic {
				isGeneric = true
				break
			}
		}

		// also treat -bin as generic to prefer the cleaner application.name
		if strings.HasSuffix(name, "-bin") {
			isGeneric = true
		}

		if !isGeneric {
			return name
		}
	}

	// 3. if it's generic/missing, try the application name (e.g. "Spotify", "Firefox")
	appName, ok := props["application.name"]
	if ok {
		return appName.String()
	}

	// 4. try the icon name
	iconName, ok := props["application.icon_name"]
	if ok {
		return iconName.String()
	}

	// 5. fallback to binary name if we have it
	if binaryName, ok := props["application.process.binary"]; ok {
		return binaryName.String()
	}

	return ""
}
