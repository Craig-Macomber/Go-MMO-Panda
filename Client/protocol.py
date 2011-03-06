from twisted.internet.protocol import Protocol, ReconnectingClientFactory
from twisted.internet import reactor, ssl
from sys import stdout
from OpenSSL import SSL

import struct

outEventTypeStruct=struct.Struct("<L")

import clientNodes

class Server(object):
    def __init__(self,name,address,port,protocol,key):
        self.name=name
        self.address=address
        self.port=port
        self.protocol=protocol
        self.key=key
        print "Server Loaded:",name,address,port,key
    def httpAddress(self,location=""):
        return self.address+":"+str(self.port)+"/"+location


class _MessageStreamer(Protocol):
    def __init__(self,owner):
        self.owner=owner
        #Protocol.__init__()
        self._protocalHeader=struct.Struct("<LL")
        self._inData=""
        
    def Uint32OnesCompliment(self,v):
        return (2**32-(v)-1)
        
    def dataReceived(self, data):
        self._inData+=data
        self._handelData()
        
    def connectionMade(self):
        self.owner.connectionMade(self)
    
    def sendMessage(self,data):
        datalen=len(data)+self._protocalHeader.size
        lenCheck=self.Uint32OnesCompliment(1 + datalen)
        header=self._protocalHeader.pack(datalen,lenCheck)
        
        self.transport.write(header+data)
    
    def _handelData(self):
        maxMessageSize=10000
        maxBackLog=50000
        hardLimitBackLog=100000
        if len(self._inData)>hardLimitBackLog:
            print "Disconnected because hardLimitBackLog of messages"
            self.owner.disconnect()
            return
        if len(self._inData)>=self._protocalHeader.size:
            
            h=self._protocalHeader.unpack(self._inData[:self._protocalHeader.size])
            length=h[0]
            lengthCheck=self.Uint32OnesCompliment(h[1])-1
            if length<>lengthCheck:
                
                print "Disconnected because of invalid message length check"
                print length
                print lengthCheck
                self.owner.disconnect()
                return
            if length > maxMessageSize:
                print "Disconnected because too long message:",length
                self.owner.disconnect()
                return
            
            if len(self._inData)>=length:
                if len(self._inData)<maxBackLog:
                    self.owner.dataReceived(self._inData[self._protocalHeader.size:length])
                else:
                    print "Excessive backlog, dropping messages"
                self._inData=self._inData[length:]
                self._handelData()
                

#### Socket Nodes: Connects to a server ####

class SocketNode(clientNodes.Parent,ReconnectingClientFactory):
    """
    auto reconnects after dropped connections
    """
    def __init__(self,server,*args,**kargs):
        """
        sendMessage should take a string and a boolean (queue) that
        indicates if a message should be dropped or queued when not connected
        """
        
        kargs["headFlag"]=0 # all sockets can have the same flag, so it might as well be 0
        
        
        self.maxDelay=5
        self.initialDelay=0.5
        self.connector=None
        self.messageStreamer=None
        self.messageQueue=[]
        
        clientNodes.Parent.__init__(self,*args,**kargs)
        
        if server.protocol=="rawTCP":
            reactor.connectTCP(server.address, int(server.port), self)
        elif server.protocol=="tlsRawTCP":
            connector = reactor.connectSSL(server.address, int(server.port), fact, ClientTLSContext())
        else:
            print 'Unsupported protocol "'+server.protocol+'" failed to connect to "'+server.name+'"'
        
        
        
        
        

    def sendEvent(self,type,data,queue=True):
        """
        sends an event back to the server
        """
        self.sendMessage(outEventTypeStruct.pack(type)+data,queue)
    
    def streamError(self): self.child.streamError() #Generally for lost or corrupt data removed at higher level
    def connected(self):pass # for initial connections and reconnects.
    




        
    def dataReceived(self, data):
        self.resetDelay() # clear the delay after sucessfully getting a message
        self.handleMessage(data)
    
    def disconnect(self):
        """ disconnect, which will lead to auto reconnect unless prevented"""
        self.connector.disconnect()
    
    def startedConnecting(self, connector):
        print 'Started to connect.'
        self.connector=connector
        
    def buildProtocol(self, addr):
        print 'Connected.'
        return _MessageStreamer(self)

    def clientConnectionLost(self, connector, reason):
        self.messageStreamer=None
        print 'Lost connection.  Reason:', reason
        self.retry()

    def clientConnectionFailed(self, connector, reason):
        self.messageStreamer=None
        print 'Connection failed. Reason:', reason
        self.retry()
    
    def connectionMade(self, messageStreamer):
        self.messageStreamer=messageStreamer
        self.connected()
        self.processQueue()
    
    def sendMessage(self,data,queue=False):
        """
        if queue, messages will be stored if not connected and sent upon connection
        dropping messages is not an error
        """
        if queue:
            self.messageQueue.append(data)
            self.processQueue()
        else:
            self.processQueue() # be sure to keep things in order
            sendMessages([data])
            
    def processQueue(self):
        failed=self.sendMessages(self.messageQueue)
        if not failed: self.messageQueue=[]
        
    def sendMessages(self,messages):
        """ returns true on fail (aka, not connected) """
        if self.messageStreamer is not None:
            for data in messages:
                self.messageStreamer.sendMessage(data)
            return False
        else:
            return True


class ClientTLSContext(ssl.ClientContextFactory):
    isClient = 1
    def getContext(self):
        #return SSL.Context(SSL.TLSv1_METHOD)
        self.method = SSL.SSLv23_METHOD
        ctx = ssl.ClientContextFactory.getContext(self)
        ctx.use_certificate_file('../Server/cert/cert.pem')
        #ctx.use_privatekey_file('../Server/cert/private.pem')

        return ctx