from twisted.internet.protocol import Protocol, ReconnectingClientFactory
from twisted.internet import reactor
from sys import stdout

import struct

# class Echo(Protocol):
#     def dataReceived(self, data):
#         #stdout.write(data)
#         print data
# 
# class EchoClientFactory(ClientFactory):
#     def startedConnecting(self, connector):
#         print 'Started to connect.'
# 
#     def buildProtocol(self, addr):
#         print 'Connected.'
#         return Echo()
# 
#     def clientConnectionLost(self, connector, reason):
#         print 'Lost connection.  Reason:', reason
# 
#     def clientConnectionFailed(self, connector, reason):
#         print 'Connection failed. Reason:', reason

# class MessageStream(Protocol):
#     def __init__(self,owner):
#         self.owner=owner
#         #Protocol.__init__()
#     def dataReceived(self, data):
#         self.owner.dataReceived(data)


class MessageStream(Protocol):
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
        
    def _handelData(self):
        if len(self._inData)>=self._protocalHeader.size:
            
            h=self._protocalHeader.unpack(self._inData[:self._protocalHeader.size])
            length=h[0]
            lengthCheck=self.Uint32OnesCompliment(h[1])-1
            if length<>lengthCheck:
                
                print "Disconnected because of invalid message length check"
                print length
                print lengthCheck
                self.disconnect()
                return
                
            if len(self._inData)>=length:
                self.owner.dataReceived(self._inData[self._protocalHeader.size:length])
                self._inData=self._inData[length:]
                self._handelData()
    
class MessageStreamClientFactory(ReconnectingClientFactory):
    def __init__(self,out):
        self.out=out
        #super(MessageStreamClientFactory,self).__init__()
        
    def dataReceived(self, data):
        self.out.gotMessage(data)
    
    def startedConnecting(self, connector):
        print 'Started to connect.'

    def buildProtocol(self, addr):
        print 'Connected.'
        return MessageStream(self)

    def clientConnectionLost(self, connector, reason):
        print 'Lost connection.  Reason:', reason
        ReconnectingClientFactory.clientConnectionLost(self, connector, reason)

    def clientConnectionFailed(self, connector, reason):
        print 'Connection failed. Reason:', reason
        ReconnectingClientFactory.clientConnectionLost(self, connector, reason)

TCPProtocolFactoryMap={
    "rawTCP":MessageStreamClientFactory,
    }

def serverToMessageStream(server,out):
    protocol=server.protocal
    if protocol in TCPProtocolFactoryMap:
        fact=TCPProtocolFactoryMap[protocol](out)
        connector = reactor.connectTCP(server.address, int(server.port), fact)
        print connector
        #reactor.connectTCP("127.0.0.1", 6666, EchoClientFactory())
    else:
        print 'Unsupported Protocal "'+protocal+'" failed to connect to "'+server.name+'"'