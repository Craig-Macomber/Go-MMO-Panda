import urllib
import struct

import pack

import protocol

outEventTypeStruct=struct.Struct("<L")

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
        

class MessageStream(object):
    def __init__(self,handler=lambda data: None):
        self.handler=handler
    def gotMessage(self,data):
        self.handler(data)

def makeStreamMultiplexer(headerLength,handlerMap):
    def handle(message):
        handlerMap[message[:headerLength]](message[headerLength:])
    return MessageStream(handle)

 

class Parent(Node):
    """ Branch of the tree, stream in, streams to children out, abstract """
    def load(self):
        self.headFlag=self.linkFields["headFlag"]
    
    def handleMessage(self,message):
        head=ord(message[0])
        if head==self.headFlag: #To me, not target
            self.handelCommand(message[1:])
        else: #send to target
            self.issueMessage(message)
            
    def handelCommand(self,message): pass #overload me
    def issueMessage(self,message): pass #overload me
    def streamError(self): pass #overload me. Generally for lost or corrupt data removed at higher level
    
class ListManager(Parent):
    """ Branch of the tree, stream in, streams to children out """
    #modes
    fullInit=0
    noneStarted=-1
    updating=1
    streamError=-2
    def load(self):
        Parent.load(self)
        self.factory=self.linkFields["factory"]
        self.remover=self.linkFields["remover"]
        self.children=[]
        self.targetIndex=-1
        self.mode=noneStarted
        self.childrenUpToDate=False
        
    def handelCommand(self,message):
        if self.mode==ListManager.fullInit:
            # full init must be done as we got a new command, so loaded
            
            #remove excess children
            i=len(self.children)-1
            while i>self.targetIndex:
                self.remover(self.children.pop())
            self.childrenUpToDate=True
            if not self.isLoaded: self.loaded()
            
            
        command=ord(message[1])
        if command==0: #empty
            for c in self.children:
                self.remover(c)
            self.children=[]
        elif command==1: #full init all
            self.mode=ListManager.fullInit
            self.targetIndex=0
        elif command==2: #updates
            self.mode=ListManager.updating
            self.targetIndex=0
        elif command==3: #swap remove
            if self.childrenUpToDate:
                self.children[self.targetIndex]=self.children.pop()
        
    def issueMessage(self,message):
        if self.mode==ListManager.fullInit:
            if self.targetIndex>=len(self.children):
                self.children.append(self.factory())
            self.children[self.targetIndex].handleMessage(message)
            self.targetIndex+=1
        elif self.mode==ListManager.updating:
            if self.childrenUpToDate:
                self.children[self.targetIndex].handleMessage(message)
                self.targetIndex+=1
            
    def streamError(self):
        self.mode=ListManager.streamError
        self.childrenUpToDate=False
            
    

#### Socket Nodes: Connects to a server ####

class SocketNode(Parent):
    """ Abstract class or origin of a MessageStream chains. A Socket Node """
    def __init__(self,server,child=None,*args,**kargs):
        """
        sendMessage should take a string and a boolean (queue) that
        indicates if a message should be dropped or queued when not connected
        """
        self.messageStream=MessageStream(self.handleMessage)
        self._sendMessage=protocol.serverToMessageStream(server,self.messageStream)
        
        kargs["headFlag"]=0 # all sockets can have the same flag, so it might as well be 0
        kargs["messageStream"]=self.messageStream
        self.child=child
        
        
        
        Parent.__init__(self,*args,**kargs)

    def sendEvent(self,type,data,queue=True):
        """
        only valid if constructed with a sendMessage
        sends an event back to the server
        """
        self._sendMessage(outEventTypeStruct.pack(type)+data,queue)
    
    def syncBack(self,data):
        """
        only valid if constructed with a sendMessage
        syncs the data back to the server to update it server side
        """
        self.sendEvent(0,data,False) # TODO : just sending sync as raw data event type 0. Could be more adaptive or more clear
    

    def handelCommand(self,message): pass #overload me
    def issueMessage(self,message):
        if self.child: self.child.handleMessage(message)
    def streamError(self): self.child.streamError() #Generally for lost or corrupt data removed at higher level
    
    
    



class HttpFile(Node):
    def load(self,skipLoaded=False):
        self.fileText=urllib.urlopen(self.linkFields["address"]).read()
        self.loaded()

## RamSync Leaf Nodes ##

class RamSync(Node):
    def load(self):
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
        head=ord(message[0])
        delta=message[1:]
        
        
        needsData=head not in pack.notNeedsData
        # if not loaded, ignore data that requires a keyframe
        if not self.isLoaded and needsData: return
        
        #apply delta here
        data=pack.binDeltaMap[head](self.data if needsData else None,delta)
        RamSync.handleMessage(self,data)

class PeriodicRequest(RamSync):
    """Asks for updates Periodically"""
    pass