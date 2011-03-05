# from twisted.internet import pollreactor
# pollreactor.install()

from twisted.internet import reactor

import clientNodes
import protocol

class Root(clientNodes.Node):
    def load(self):
        print "Loading Root"
        
        # hard coded root data
        self.masterPublicKey="xxx"
        self.masterServerAddress="http://127.0.0.1"
        
        # load subnode, the serverList
        masterServer=protocol.Server("master",self.masterServerAddress,8080,"http",self.masterPublicKey)
        
        self.serverList=ServerList(address=masterServer.httpAddress("ServerList"),
            key=masterServer.key,
            loadCallBacks=[self.loaded])

    def loaded(self):
        clientNodes.Node.loaded(self)
        print "root loaded"

class ServerList(clientNodes.HttpFile):
    def load(self):
        print "Loading Server List"
        clientNodes.HttpFile.load(self,skipLoaded=False)
        
        # if read failed, should report it here, and use version from local cache
        
        key=self.linkFields["key"]
        
        # TODO: Verify listText is properly signed here
       
        # parse listText
        lines=self.fileText.splitlines()
        self.servers={}
        for s in lines[1:]:
            t=s.split()
            self.servers[t[1]]=protocol.Server(*t[1:6])
        
        self.loaded()
        print "ServerList loaded"
        
class TestServer(clientNodes.KeyFrameBinDelta):
    def load(self): 
        clientNodes.KeyFrameBinDelta.load(self)
        self.updatedCallbacks.append(self.handelUpdate)
    def handelUpdate(self):
        pass#print "Got Update: "#+self.data


# need to make login not try and auto reconnect if login fails (perhaps some sort of time out too)
# and login after disconnects if login suceeded
class LoggedInSocket(clientNodes.SocketNode):
    def load(self): 
        clientNodes.SocketNode.load(self)
        
        userName="Test"
        password="12345"
        self.sendEvent(1,userName)
        self.sendEvent(2,password)

        self.loggedIn=False
        
    def handelCommand(self,message):
        head=ord(message[0])
        self.loggedIn=head==0 # 0 means login sucess
        print "login",self.loggedIn
        self.loaded()
            

root=Root()
x=[]
for i in range(1):
    testServer=TestServer()
    server=root.serverList.servers["Login"]
    socketNode=LoggedInSocket(server,testServer)
    x.append(testServer)
    #socketNode.sendEvent(1,"Test")

reactor.run()