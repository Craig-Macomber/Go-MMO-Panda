from twisted.internet.protocol import Protocol, ReconnectingClientFactory
from twisted.internet import reactor
from sys import stdout

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

class MessageStream(Protocol):
    def __init__(self,owner):
        self.owner=owner
        #Protocol.__init__()
    def dataReceived(self, data):
        self.owner.dataReceived(data)

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

    def clientConnectionFailed(self, connector, reason):
        print 'Connection failed. Reason:', reason


TCPProtocolFactoryMap={
    "rawTCP":MessageStreamClientFactory,
    }

def serverToMessageStream(server,out):
    protocol=server.protocal
    if protocol in TCPProtocolFactoryMap:
        fact=TCPProtocolFactoryMap[protocol](out)
        reactor.connectTCP(server.address, int(server.port), fact)
        #reactor.connectTCP("127.0.0.1", 6666, EchoClientFactory())
    else:
        print 'Unsupported Protocal "'+protocal+'" failed to connect to "'+server.name+'"'