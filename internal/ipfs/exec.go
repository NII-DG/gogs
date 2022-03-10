package ipfs

//"os/exec"
//"github.com/ipfs/go-ipfs-api"

type IpfsCommand struct {
	name string
	args []string
	envs []string
}

func Execution(args ...string) *IpfsCommand {
	return &IpfsCommand{
		name: "ipfs",
		args: args,
	}
}
