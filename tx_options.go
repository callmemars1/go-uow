package uow

type IsolationLevel string

const (
	ReadUncommitted IsolationLevel = "read uncommitted"
	ReadCommitted   IsolationLevel = "read committed"
	RepeatableRead  IsolationLevel = "repeatable read"
	Serializable    IsolationLevel = "serializable"
)

type TxOptions struct {
	IsolationLevel IsolationLevel
	ReadOnly       bool
}

func DefaultTxOptions() TxOptions {
	return TxOptions{
		IsolationLevel: ReadCommitted,
		ReadOnly:       false,
	}
}

func SerializableTxOptions() TxOptions {
	return TxOptions{
		IsolationLevel: Serializable,
		ReadOnly:       false,
	}
}

func ReadOnlyTxOptions() TxOptions {
	return TxOptions{
		IsolationLevel: ReadCommitted,
		ReadOnly:       true,
	}
}
