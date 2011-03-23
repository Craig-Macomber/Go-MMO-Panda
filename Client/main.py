# from twisted.internet import pollreactor
# pollreactor.install()

from twisted.internet import reactor
from twisted.internet.task import LoopingCall

import clientNodes
import protocol
import nodes

class Root(clientNodes.Node):
    def load(self):
        print "Loading Root"
        
        # hard coded root data
        self.masterPublicKey="xxx"
        self.masterServerAddress="http://127.0.0.1"
        
        # load subnode, the serverList
        masterServer=protocol.Server("master",self.masterServerAddress,8080,"http",self.masterPublicKey)
        
        self.serverList=nodes.ServerList(address=masterServer.httpAddress("ServerList"),
            key=masterServer.key,
            loadCallBacks=[self.loaded])

    def loaded(self):
        clientNodes.Node.loaded(self)
        print "root loaded"

        
class TestServer(clientNodes.KeyFrameBinDelta):
    def load(self): 
        clientNodes.KeyFrameBinDelta.load(self)
        self.updatedCallbacks.append(self.handelUpdate)
    def handelUpdate(self):
        pass #print "Got Update: ",len(self.data),self.data

class EventCatcher(clientNodes.Node):
    def handelMessage(self,message):
        data=message[1:]
        print "Event: "+data



root=Root()
x=[]
for i in range(1):
    testServer=TestServer()
    eventCatcher=EventCatcher()
    server=root.serverList.servers["Login"]
    socketNode=nodes.LoggedInSocket(server)
    socketNode.addChild(testServer,1)
    socketNode.addChild(eventCatcher,2)
    x.append(socketNode)


socketNode=x[0]

# init panda3d
import direct.showbase.ShowBase
base=direct.showbase.ShowBase.ShowBase()
def testFunc():
    socketNode.sendEvent(3,"MessageTest")
    
base.accept("a",testFunc)

LoopingCall(taskMgr.step).start(1.0 / 60)
reactor.run()