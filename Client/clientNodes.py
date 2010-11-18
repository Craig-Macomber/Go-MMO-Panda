import urllib
import struct

import pack

class Node(object):
    def __init__(self,loadCallbacks=None,**linkFields):
        self.linkFields=linkFields # How to connect to this node
        
        # cleared on load, thus per load
        self.loadCallBacks=[] if loadCallbacks is None else loadCallbacks
        
        self.forceUpdate()

    def forceUpdate(self):
        self.isLoaded=False
        self.load()
    
    def loaded(self):
        self.isLoaded=True
        for c in self.loadCallBacks: c()
        self.loadCallBacks=[]
    
    def load(self): pass #Overload Me        
        
class HttpFile(Node):        
    def load(self,skipLoaded=False):
        self.fileText=urllib.urlopen(self.linkFields["address"]).read()
        self.loaded()


class MessageStream(object):
    def __init__(self):
        self.handler=lambda data: None #throw away unhandled data
    def gotMessage(self,data):
        self.handler(data)

def makeStreamMultiplexer(headerLength,handlerMap):
        def handle(message):
            handlerMap[message[:headerLength]](message[headerLength:])
        return handle
    

class RamSync(Node):
    def load(self):
        self.stream=self.linkFields["messageStream"]
        self.stream.handler=self.handleMessage
        self.updatedCallbacks=[]
        
    def handleMessage(self,message):
        self.data=message
        if not self.isLoaded: self.loaded()
        self.updated()
        
    def updated(self):
        for f in self.updatedCallbacks: f()

class StructNode(RamSync):
    def load(self):
        self.struct=struct.Struct(self.linkFields["format"])
        RamSync.load(self)
    def unpack(self):
        return self.struct.unpack(self.data)

class KeyFrameBinDelta(RamSync):
    """ Syncs data based on key frames containing the whole data, and some kind of diff packet """
    def handleMessage(self,message):
        headerLength=1
        head=ord(message[:headerLength])
        delta=message[headerLength:]
        
        
        needsData=head not in (0,1)
        # if not loaded, ignore data that requires a keyframe
        if not self.isLoaded and needsData: return
        
        #apply delta here
        data=pack.binDeltaMap[head](self.data if needsData else None,delta)
        RamSync.handleMessage(self,data)

class PeriodicRequest(RamSync):
    """Asks for updates Periodically"""
    pass