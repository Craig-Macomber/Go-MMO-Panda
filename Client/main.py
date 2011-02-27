from twisted.internet import reactor

from clientNodes import Node, HttpFile, MessageStream, KeyFrameBinDelta, RamSync
import protocol

class Root(Node):
    def load(self):
        print "Loading Root"
        
        # hard coded root data
        self.masterPublicKey="xxx"
        self.masterServerAddress="http://127.0.0.1"
        
        # load subnode, the serverList
        masterServer=Server("master",self.masterServerAddress,8080,"http",self.masterPublicKey)
        
        self.serverList=ServerList(address=masterServer.httpAddress("ServerList"),
            key=masterServer.key,
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
        print lines
        self.servers={}
        for s in lines[1:]:
            t=s.split()
            self.servers[t[1]]=Server(*t[1:6])
        
        self.loaded()
        print "ServerList loaded"
        
class Server(object):
    def __init__(self,name,address,port,protocal,key):
        self.name=name
        self.address=address
        self.port=port
        self.protocal=protocal
        self.key=key
        print "Server Loaded:",name,address,port,key
    def httpAddress(self,location=""):
        return self.address+":"+str(self.port)+"/"+location
        
class TestServer(KeyFrameBinDelta):
    def load(self): 
        KeyFrameBinDelta.load(self)
        self.updatedCallbacks.append(self.handelUpdate)
    def handelUpdate(self):
        pass#print "Got Update: "#+self.data

root=Root()
x=[]
for i in range(3):
    stream=MessageStream()
    testServer=TestServer(messageStream=stream)
    protocol.serverToMessageStream(root.serverList.servers["Login"],stream)
    x.append(testServer)


reactor.run()