package prompt

// SetDescriptions is a create run's memory of what each set was described as. The
// description belongs to the set, not the photograph, so a batch is asked once per
// set and later files of that set are offered the same answer.
type SetDescriptions map[string]string

func NewSetDescriptions() SetDescriptions {
	return SetDescriptions{}
}

// Default is what the set was described as earlier in this run, or "" when the run
// has not described it yet.
func (d SetDescriptions) Default(set string) string {
	return d[set]
}

// Record keeps the answer that will be written, so the next file of the set is
// offered it. An empty description is not recorded: it says nothing about the set.
func (d SetDescriptions) Record(set, description string) {
	if description != "" {
		d[set] = description
	}
}
