from twisted.internet import reactor

from clientNodes import Node, HttpFile, MessageStream, KeyFrameBinDelta
import protocol

class Root(Node):
    def load(self):
        print "Loading Root"
        
        # hard coded root data
        self.masterPublicKey="xxx"
        self.masterServerAddress="http://127.0.0.1:8080/serverList"
        
        # load subnode, the serverList
        self.serverList=ServerList(
                    key=self.masterPublicKey,
                    address=self.masterServerAddress,
                    loadCallBacks=[self.loaded])

    def loaded(self):
        Node.loaded(self)
        print "root loaded"

class ServerList(HttpFile):
    def load(self):
        print "Loading Server List"
        HttpFile.load(self,skipLoaded=False)
        
        # if read failed, should report it here, and use version from local cache
        
        key=self.linkFields["key"]
        
        # TODO: Verify listText is properly signed here
       
        # parse listText
        lines=self.fileText.splitlines()
        self.servers={}
        for s in lines[1:]:
            t=s.split()
            self.servers[t[0]]=Server(*t[:4])
        
        self.loaded()
        print "ServerList loaded"
        
class Server(object):
    def __init__(self,name,address,port,protocal):
        self.name=name
        self.address=address
        self.port=port
        self.protocal=protocal
    def getConnection(self,params=None):
        if self.protocal=='fgTCP':
            print "Not supported!"
        else:
            print 'Unsupported Protocal "'+self.protocal+'" failed to connect to "'+self.name+'"'

class TestServer(KeyFrameBinDelta):
    def load(self): 
        KeyFrameBinDelta.load(self)
        self.updatedCallbacks.append(self.handelUpdate)
    def handelUpdate(self):
        print "Test Update"
        print self.data

root=Root()
stream=MessageStream()
testServer=TestServer(messageStream=stream)
protocol.serverToMessageStream(root.serverList.servers["Login"],stream)


reactor.run()