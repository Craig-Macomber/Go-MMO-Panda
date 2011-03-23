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
    def streamError(self): pass
 

class Parent(Node):
    """ Branch of the tree, stream in, streams to children out
    
    unless overwritten, streams messages to children based on
    first byte of message, headFlag
    redirects message stream when ever a recognized headFlag comes by.
    """
    
    
    def load(self):
        self.headFlag=self.linkFields["headFlag"]
        self.children={}
        self.lastFlag=self.headFlag
        
    def addChild(self,node,headFlag):
        self.children[headFlag]=node
    
    def handleMessage(self,message):
        head=ord(message[0])
        if head==self.headFlag: #To me, not target
            self.lastFlag=self.headFlag
            self.handelCommand(message[1:])
        else: #send to target
            self.issueMessage(message)
            
    def handelCommand(self,message): pass #overload me
    def issueMessage(self,message):
        head=ord(message[0])
        if head in self.children:
            self.lastFlag=head
            self.children[head].handleMessage(message)
        else:
            if self.lastFlag!=self.headFlag:
                self.children[lastFlag].handleMessage(message)
            else:
                print "Dropping message with no target",head
        
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
            
    


class HttpFile(Node):
    def load(self,skipLoaded=False):
        self.fileText=urllib.urlopen(self.linkFields["address"]).read()
        self.loaded()

## RamSync Leaf Nodes ##

class RamSync(Node):
    def load(self):
        self.updatedCallbacks=[]
        self.hasSync=False
        
    def handleMessage(self,message): self.handelCommand(message[1:])
    def handleCommand(self,message):
        self.data=message
        self.hasSync=True
        if not self.isLoaded: self.loaded()
        self.updated()
        
    def updated(self):
        for f in self.updatedCallbacks: f()
    
    def streamError(self): self.hasSync=False
    
    
class StructNode(RamSync):
    def load(self):
        self.struct=struct.Struct(self.linkFields["format"])
        RamSync.load(self)
    def unpack(self):
        return self.struct.unpack(self.data)

class KeyFrameBinDelta(RamSync):
    """ Syncs data based on key frames containing the whole data, and some kind of diff packet """
    def handleMessage(self,message):
        headFlag=ord(message[0])
        head=ord(message[1])
        delta=message[2:]
        
        
        needsData=head not in pack.notNeedsData
        # if not loaded, ignore data that requires a keyframe
        if not self.hasSync and needsData: return
        
        #apply delta here
        data=pack.binDeltaMap[head](self.data if needsData else None,delta)
        if data is None:
            self.hasSync=False
        else:
            RamSync.handleCommand(self,data)

class PeriodicRequest(RamSync):
    """Asks for updates Periodically"""
    pass