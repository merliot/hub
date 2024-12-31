package device

type deviceMap map[string]*device // key: device id

var devices = make(deviceMap)
var devicesMu rwMutex
