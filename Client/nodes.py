import clientNodes
import protocol

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



# need to make login not try and auto reconnect if login fails (perhaps some sort of time out too)
# and login after disconnects if login suceeded
class LoggedInSocket(protocol.SocketNode):
    """
    super insecure for now
    """
    def load(self): 
        protocol.SocketNode.load(self)
        
        self.userName="Test"
        self.password="12345"
        

        self.loggedIn=False
    
    def connected(self):
        self.sendEvent(1,self.userName)
        self.sendEvent(2,self.password)
        self.loggedIn=False
        
    def handelCommand(self,message):
        head=ord(message[0])
        self.loggedIn=head==0 # 0 means login sucess
        print "login",self.loggedIn
        self.loaded()