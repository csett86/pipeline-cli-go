package cli

import (
	"fmt"
	"os"
	"text/template"

	"github.com/daisy-consortium/pipeline-clientlib-go"
)

const (
	JobStatusTemplate = `
Job Id: {{.Data.Id }}
Status: {{.Data.Status}}
{{if .Verbose}}Messages:
{{range .Data.Messages}}
({{.Sequence}})[{{.Level}}]      {{.Content}}
{{end}}
{{end}}
`

	JobListTemplate = `
Job Id          (Nicename)              [STATUS]

{{range .}}{{.Id}}{{if .Nicename }}     ({{.Nicename}}){{end}}  [{{.Status}}]
{{end}}`

	VersionTemplate = `
Client version:                 {{.CliVersion}}         
Pipeline version:               {{.Version}}
Pipeline authentication:        {{.Authentication}}
`

	QueueTemplate = `Job Id 			Priority	Job P.	 Client P.	Rel.Time.	 Since
{{range .}}{{.Id}}	{{.ComputedPriority | printf "%.2f"}}	{{.JobPriority}}	{{.ClientPriority}}	{{.RelativeTime | printf "%.2f"}}	{{.TimeStamp}}
{{end}}`
)

//Convinience struct
type printableJob struct {
	Data    pipeline.Job
	Verbose bool
}

func AddJobStatusCommand(cli *Cli, link PipelineLink) {
	printable := &printableJob{
		Data:    pipeline.Job{},
		Verbose: false,
	}
	fn := func(args ...interface{}) (interface{}, error) {
		job, err := link.Job(args[0].(string))
		if err != nil {
			return nil, err
		}
		printable.Data = job
		return printable, nil
	}
	cmd := newCommandBuilder("status", "Returns the status of the job with id JOB_ID").
		withCall(fn).withTemplate(JobStatusTemplate).
		buildWithId(cli)

	cmd.AddSwitch("verbose", "v", "Prints the job's messages", func(swtich, nop string) error {
		printable.Verbose = true
		return nil
	})
}

func AddDeleteCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	cmd := cli.AddCommand("delete", "Removes a job from the pipeline", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		ok, err := link.Delete(id)
		if err != nil {
			return err
		}
		if ok {
			fmt.Printf("Job %v removed\n", id)
		}

		return nil
	})
	addLastId(cmd, lastId)
}

func AddResultsCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	outputPath := ""
	cmd := cli.AddCommand("results", "Stores the results from a job", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		data, err := link.Results(id)
		if err != nil {
			return err
		}

		path, err := zippedDataToFolder(data, outputPath)
		if err != nil {
			return err
		}

		fmt.Printf("Results stored into %v\n", path)

		return nil
	})
	cmd.AddOption("output", "o", "Directory where to store the results", func(name, folder string) error {
		outputPath = folder
		return nil
	}).Must(true)
	addLastId(cmd, lastId)
}

func AddLogCommand(cli *Cli, link PipelineLink) {
	lastId := new(bool)
	outputPath := ""
	cmd := cli.AddCommand("log", "Stores the results from a job", func(command string, args ...string) error {
		id, err := checkId(*lastId, command, args...)
		if err != nil {
			return err
		}
		data, err := link.Log(id)
		if err != nil {
			return err
		}
		outWriter := os.Stdout
		if len(outputPath) > 0 {
			file, err := os.Create(outputPath)
			defer func() { file.Close() }()
			if err != nil {
				return err
			}
			outWriter = file
		}
		_, err = outWriter.Write(data)
		if err != nil {
			return err
		}
		return nil
	})
	cmd.AddOption("output", "o", "Write the log lines into the file provided instead of printing it", func(name, file string) error {
		outputPath = file
		return nil
	})
	addLastId(cmd, lastId)
}
func AddHaltCommand(cli *Cli, link PipelineLink) {
	cli.AddCommand("halt", "Stops the webservice", func(command string, args ...string) error {
		key, err := loadKey()
		if err != nil {
			return err
		}
		err = link.Halt(key)
		if err != nil {
			return err
		}
		fmt.Println("The webservice has been halted")
		return nil
	})
}

func AddJobsCommand(cli *Cli, link PipelineLink) {
	cli.AddCommand("jobs", "Returns the list of jobs present in the server", func(command string, args ...string) error {
		jobs, err := link.Jobs()
		if err != nil {
			return err
		}
		tmpl, err := template.New("joblist").Parse(JobListTemplate)
		if err != nil {
			return err
		}
		err = tmpl.Execute(os.Stdout, jobs)
		return nil
	})
}

func AddQueueCommand(cli *Cli, link PipelineLink) {
	fn := func(...interface{}) (queue interface{}, err error) {
		queue, err = link.Queue()
		return
	}
	newCommandBuilder("queue", "Shows the execution queue and the job's priorities. ").
		withCall(fn).withTemplate(QueueTemplate).build(cli)
}

func AddMoveUpCommand(cli *Cli, link PipelineLink) {
	fn := func(args ...interface{}) (queue interface{}, err error) {
		queue, err = link.MoveUp(args[0].(string))
		return
	}
	newCommandBuilder("moveup", "Moves the job up the execution queue").
		withCall(fn).withTemplate(QueueTemplate).
		buildWithId(cli)

}

type Version struct {
	*PipelineLink
	CliVersion string
}

func AddVersionCommand(cli *Cli, link *PipelineLink) {
	newCommandBuilder("version", "Prints the version and authentication information").
		withCall(func(...interface{}) (interface{}, error) {
		return Version{link, VERSION}, nil
	}).withTemplate(VersionTemplate).build(cli)

}
