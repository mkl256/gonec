package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/covrom/gonec/names"
)

const (
	VMNanosecond  VMTimeDuration = 1
	VMMicrosecond                = 1000 * VMNanosecond
	VMMillisecond                = 1000 * VMMicrosecond
	VMSecond                     = 1000 * VMMillisecond
	VMMinute                     = 60 * VMSecond
	VMHour                       = 60 * VMMinute
	VMDay                        = 24 * VMHour
)

// VMTimeDuration - диапазон между отметками времени

type VMTimeDuration time.Duration

var ReflectVMTimeDuration = reflect.TypeOf(VMTimeDuration(0))

func (v VMTimeDuration) vmval() {}

func (v VMTimeDuration) Interface() interface{} {
	return time.Duration(v)
}

func (v VMTimeDuration) Duration() VMTimeDuration {
	return v
}

func (x VMTimeDuration) BinaryType() VMBinaryType {
	return VMDURATION
}

func (v VMTimeDuration) String() string {
	u := uint64(v)
	if u == 0 {
		return "0с"
	}
	neg := v < 0
	if neg {
		u = -u
	}

	var buf [40]byte
	w := len(buf)

	if u < uint64(VMSecond) {
		// Special case: if duration is smaller than a second,
		// use smaller units, like 1.2ms
		var prec int
		w -= 2
		copy(buf[w:], "с")
		switch {
		case u < uint64(VMMicrosecond):
			// print nanoseconds
			prec = 0
			w -= 2 // Need room for two bytes.
			copy(buf[w:], "н")
		case u < uint64(VMMillisecond):
			// print microseconds
			prec = 3
			w -= 4 // Need room for 4 bytes.
			copy(buf[w:], "мк")
		default:
			// print milliseconds
			prec = 6
			w -= 2 // Need room for two bytes.
			copy(buf[w:], "м")
		}
		w, u = fmtFrac(buf[:w], u, prec)
		w = fmtInt(buf[:w], u)
	} else {
		w -= 2
		copy(buf[w:], "с")

		w, u = fmtFrac(buf[:w], u, 9)

		// u is now integer seconds
		w = fmtInt(buf[:w], u%60)
		u /= 60

		// u is now integer minutes
		if u > 0 {
			w -= 2
			copy(buf[w:], "м")
			w = fmtInt(buf[:w], u%60)
			u /= 60

			// u is now integer hours
			if u > 0 {
				w -= 2
				copy(buf[w:], "ч")
				w = fmtInt(buf[:w], u%24)
				u /= 24

				if u > 0 {
					w -= 2
					copy(buf[w:], "д")
					w = fmtInt(buf[:w], u)
				}
			}
		}
	}

	if neg {
		w--
		buf[w] = '-'
	}

	return string(buf[w:])
}

func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	// Omit trailing zeros up to and including decimal point.
	w := len(buf)
	print := false
	for i := 0; i < prec; i++ {
		digit := v % 10
		print = print || digit != 0
		if print {
			w--
			buf[w] = byte(digit) + '0'
		}
		v /= 10
	}
	if print {
		w--
		buf[w] = '.'
	}
	return w, v
}

func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

func (x VMTimeDuration) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, int64(x))
	return buf.Bytes(), nil
}

func (x *VMTimeDuration) UnmarshalBinary(data []byte) error {
	var i int64
	if err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &i); err != nil {
		return err
	}
	*x = VMTimeDuration(i)
	return nil
}

func (x VMTimeDuration) GobEncode() ([]byte, error) {
	return x.MarshalBinary()
}

func (x *VMTimeDuration) GobDecode(data []byte) error {
	return x.UnmarshalBinary(data)
}

func (x VMTimeDuration) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(time.Duration(x).String())
	return buf.Bytes(), nil
}

func (x *VMTimeDuration) UnmarshalText(data []byte) error {
	d, err := time.ParseDuration(string(data))
	if err != nil {
		return err
	}
	*x = VMTimeDuration(d)
	return nil
}

func (x VMTimeDuration) MarshalJSON() ([]byte, error) {
	b, err := x.MarshalText()
	if err != nil {
		return nil, err
	}
	return []byte("\"" + string(b) + "\""), nil
}

func (x *VMTimeDuration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if len(data) > 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}
	return x.UnmarshalText(data)
}

//VMTime дата и время

type VMTime time.Time

var ReflectVMTime = reflect.TypeOf(VMTime{})

func Now() VMTime {
	return VMTime(time.Now())
}

func (v VMTime) vmval() {}

func (v VMTime) Interface() interface{} {
	return time.Time(v)
}

func (t VMTime) Time() VMTime {
	return t
}

func (x VMTime) BinaryType() VMBinaryType {
	return VMTIME
}

func (t VMTime) MarshalBinary() ([]byte, error) {
	return time.Time(t).MarshalBinary()
}

func (t *VMTime) UnmarshalBinary(data []byte) error {
	tt := time.Time(*t)
	err := (&tt).UnmarshalBinary(data)
	if err != nil {
		return err
	}
	*t = VMTime(tt)
	return nil
}

func (t VMTime) GobEncode() ([]byte, error) {
	return time.Time(t).GobEncode()
}

func (t *VMTime) GobDecode(data []byte) error {
	tt := time.Time(*t)
	err := (&tt).GobDecode(data)
	if err != nil {
		return err
	}
	*t = VMTime(tt)
	return nil
}

func (t VMTime) MarshalJSON() ([]byte, error) {
	return time.Time(t).MarshalJSON()
}

func (t *VMTime) UnmarshalJSON(data []byte) error {
	tt := time.Time(*t)
	err := (&tt).UnmarshalJSON(data)
	if err != nil {
		return err
	}
	*t = VMTime(tt)
	return nil
}

func (t VMTime) MarshalText() ([]byte, error) {
	return time.Time(t).MarshalText()
}

func (t *VMTime) UnmarshalText(data []byte) error {
	tt := time.Time(*t)
	err := (&tt).UnmarshalText(data)
	if err != nil {
		return err
	}
	*t = VMTime(tt)
	return nil
}

func (t VMTime) String() string {
	return time.Time(t).String()
}

func (t VMTime) Format(layout string) string {
	// формат в стиле Го
	const bufSize = 64
	var b []byte
	max := len(layout) + 10
	if max < bufSize {
		var buf [bufSize]byte
		b = buf[:0]
	} else {
		b = make([]byte, 0, max)
	}
	b = time.Time(t).AppendFormat(b, layout)
	return string(b)
}

func (t VMTime) MethodMember(name int) (VMFunc, bool) {

	// только эти методы будут доступны из кода на языке Гонец!

	switch names.UniqueNames.GetLowerCase(name) {
	case "год":
		return VMFuncMustParams(0, t.Год), true
	case "месяц":
		return VMFuncMustParams(0, t.Месяц), true
	case "день":
		return VMFuncMustParams(0, t.День), true
	case "неделя":
		return VMFuncMustParams(0, t.Неделя), true
	case "деньнедели":
		return VMFuncMustParams(0, t.ДеньНедели), true

		// TODO: остальные методы!!!
	}

	return nil, false
}

func (t VMTime) Год(args VMSlice, rets *VMSlice) error {
	rets.Append(VMInt(time.Time(t).Year()))
	return nil
}

func (t VMTime) Месяц(args VMSlice, rets *VMSlice) error {
	rets.Append(VMInt(time.Time(t).Month()))
	return nil
}

func (t VMTime) День(args VMSlice, rets *VMSlice) error {
	rets.Append(VMInt(time.Time(t).Day()))
	return nil
}

func (t VMTime) Weekday() int64 {
	//0=воскресенье, 1=понедельник ...
	return int64(time.Time(t).Weekday())
}

func (t VMTime) Неделя(args VMSlice, rets *VMSlice) error {
	// по ISO 8601
	// 1-53, выровнены по годам,
	// т.е. конец декабря (три дня) попадает в следующий год,
	// или первые три дня января попадают в предыдущий год
	yy, ww := time.Time(t).ISOWeek()
	rets.Append(VMInt(ww))
	rets.Append(VMInt(yy))
	return nil
}

func (t VMTime) ДеньНедели(args VMSlice, rets *VMSlice) error {
	//1=понедельник, 7=воскресенье ...
	wd := int64(time.Time(t).Weekday())
	if wd == 0 {
		rets.Append(VMInt(7))
	} else {
		rets.Append(VMInt(wd))
	}
	return nil
}

func (t VMTime) Квартал() int64 {
	//1-4
	return int64(time.Time(t).Month())/4 + 1
}

func (t VMTime) ДеньГода() int64 {
	//1-366
	return int64(time.Time(t).YearDay())
}

func (t VMTime) Час() int64 {
	return int64(time.Time(t).Hour())
}

func (t VMTime) Минута() int64 {
	return int64(time.Time(t).Minute())
}

func (t VMTime) Секунда() int64 {
	return int64(time.Time(t).Second())
}

func (t VMTime) Миллисекунда() int64 {
	return int64(time.Time(t).Nanosecond()) / 1e6
}

func (t VMTime) Микросекунда() int64 {
	return int64(time.Time(t).Nanosecond()) / 1e3
}

func (t VMTime) Наносекунда() int64 {
	return int64(time.Time(t).Nanosecond())
}

func (t VMTime) UnixNano() int64 {
	return time.Time(t).UnixNano()
}

func (t VMTime) Unix() int64 {
	return time.Time(t).Unix()
}

func (t VMTime) GolangTime() time.Time {
	return time.Time(t)
}

func (t VMTime) Формат(fmtstr string) string {

	// д (d) - день месяца (цифрами) без лидирующего нуля;
	// дд (dd) - день месяца (цифрами) с лидирующим нулем;
	// ддд (ddd) - краткое название дня недели *);
	// дддд (dddd) - полное название дня недели *);

	// М (M) - номер месяца (цифрами) без лидирующего нуля;
	// ММ (MM) - номер месяца (цифрами) с лидирующим нулем;
	// МММ (MMM) - краткое название месяца *);
	// ММММ (MMMM) - полное название месяца *);

	// К (Q) - номер квартала в году;
	// г (y) - номер года без века и лидирующего нуля;
	// гг (yy) - номер года без века с лидирующим нулем;
	// гггг (yyyy) - номер года с веком;

	// ч (h) - час в 24 часовом варианте без лидирующих нулей;
	// чч (hh) - час в 24 часовом варианте с лидирующим нулем;

	// м (m) - минута без лидирующего нуля;
	// мм (mm) - минута с лидирующим нулем;

	// с (s) - секунда без лидирующего нуля;
	// сс (ss) - секунда с лидирующим нулем;
	// ссс (sss) - миллисекунда с лидирующим нулем

	days := [...]string{
		"", //0-го не бывает
		"понедельник",
		"вторник",
		"среда",
		"четверг",
		"пятница",
		"суббота",
		"воскресенье",
	}

	months1 := [...]string{
		"", //0-го не бывает
		"январь",
		"февраль",
		"март",
		"апрель",
		"май",
		"июнь",
		"июль",
		"август",
		"сентябрь",
		"октябрь",
		"ноябрь",
		"декабрь",
	}

	months2 := [...]string{
		"", //0-го не бывает
		"января",
		"февраля",
		"марта",
		"апреля",
		"мая",
		"июня",
		"июля",
		"августа",
		"сентября",
		"октября",
		"ноября",
		"декабря",
	}

	dayssm := [...]string{
		"пн",
		"вт",
		"ср",
		"чт",
		"пт",
		"сб",
		"вс",
	}

	src := []rune(fmtstr)
	res := make([]rune, 0, len(src)*2)
	wasday := false
	hour, min, sec := time.Time(t).Clock()
	y, m, d := time.Time(t).Date()

	i := 0
	for i < len(src) {
		var s []rune

		if i+4 <= len(src) {
			s = src[i : i+4]
			switch string(s) {
			case "дддд", "dddd":
				res = append(res, []rune(days[t.ДеньНедели()])...)
				i += 4
				continue
			case "ММММ", "MMMM":
				if wasday {
					res = append(res, []rune(months2[t.Месяц()])...)
				} else {
					res = append(res, []rune(months1[t.Месяц()])...)
				}
				i += 4
				continue
			case "гггг", "yyyy":
				res = append(res, []rune(strconv.FormatInt(t.Год(), 10))...)
				i += 4
				continue

			}
		}

		if i+3 <= len(src) {
			s = src[i : i+3]
			switch string(s) {
			case "ддд", "ddd":
				res = append(res, []rune(dayssm[t.ДеньНедели()])...)
				i += 3
				continue
			case "МММ", "MMM":
				if wasday {
					res = append(res, []rune(months2[t.Месяц()])[:3]...)
				} else {
					res = append(res, []rune(months1[t.Месяц()])[:3]...)
				}
				i += 3
				continue
			case "ссс", "sss":
				sm := strconv.FormatInt(t.Миллисекунда(), 10)
				if len(sm) < 3 {
					sm = strings.Repeat("0", 3-len(sm)) + sm
				}
				res = append(res, []rune(sm)...)
				i += 3
				continue
			}
		}

		if i+2 <= len(src) {
			s = src[i : i+2]
			switch string(s) {
			case "дд", "dd":
				sm := strconv.Itoa(d)
				if len(sm) < 2 {
					sm = "0" + sm
				}
				res = append(res, []rune(sm)...)
				i += 2
				wasday = true
				continue
			case "ММ", "MM":
				sm := strconv.Itoa(int(m))
				if len(sm) < 2 {
					sm = "0" + sm
				}
				res = append(res, []rune(sm)...)
				i += 2
				continue
			case "гг", "yy":
				sm := strconv.Itoa(int(y % 100))
				if len(sm) < 2 {
					sm = "0" + sm
				}
				res = append(res, []rune(sm)...)
				i += 2
				continue
			case "чч", "hh":
				sm := strconv.Itoa(int(hour))
				if len(sm) < 2 {
					sm = "0" + sm
				}
				res = append(res, []rune(sm)...)
				i += 2
				continue
			case "мм", "mm":
				sm := strconv.Itoa(int(min))
				if len(sm) < 2 {
					sm = "0" + sm
				}
				res = append(res, []rune(sm)...)
				i += 2
				continue
			case "сс", "ss":
				sm := strconv.Itoa(int(sec))
				if len(sm) < 2 {
					sm = "0" + sm
				}
				res = append(res, []rune(sm)...)
				i += 2
				continue
			}
		}

		c := src[i]
		switch c {
		case 'д', 'd':
			sm := strconv.Itoa(d)
			res = append(res, []rune(sm)...)
			i++
			wasday = true
			continue
		case 'М', 'M':
			sm := strconv.Itoa(int(m))
			res = append(res, []rune(sm)...)
			i++
			continue
		case 'г', 'y':
			sm := strconv.Itoa(int(y % 100))
			res = append(res, []rune(sm)...)
			i++
			continue
		case 'ч', 'h':
			sm := strconv.Itoa(int(hour))
			res = append(res, []rune(sm)...)
			i++
			continue
		case 'м', 'm':
			sm := strconv.Itoa(int(min))
			res = append(res, []rune(sm)...)
			i++
			continue
		case 'с', 's':
			sm := strconv.Itoa(int(sec))
			res = append(res, []rune(sm)...)
			i++
			continue
		case 'К', 'Q':
			sm := strconv.FormatInt(t.Квартал(), 10)
			res = append(res, []rune(sm)...)
			i++
			continue
		}
		res = append(res, c)
		i++
	}

	return string(res)
}

func (t VMTime) Вычесть(t2 VMTime) VMTimeDuration {
	return VMTimeDuration(time.Time(t).Sub(time.Time(t2)))
}

func (t VMTime) Добавить(d VMTimeDuration) VMTime {
	return VMTime(time.Time(t).Add(time.Duration(d)))
}

func (t VMTime) ДобавитьПериод(dy, dm, dd int) VMTime {
	return VMTime(time.Time(t).AddDate(dy, dm, dd))
}

func (t VMTime) Раньше(d VMTime) bool {
	return time.Time(t).Before(time.Time(d))
}

func (t VMTime) Позже(d VMTime) bool {
	return time.Time(t).After(time.Time(d))
}

func (t VMTime) Равно(d VMTime) bool {
	// для разных локаций тоже работает, в отличие от =
	return time.Time(t).Equal(time.Time(d))
}

func (t VMTime) Пустая() bool {
	return time.Time(t).IsZero()
}

func (t VMTime) Местное() VMTime {
	return VMTime(time.Time(t).Local())
}

func (t VMTime) UTC() VMTime {
	return VMTime(time.Time(t).UTC())
}

func (t VMTime) Локация() string {
	return time.Time(t).Location().String()
}

func (t VMTime) ВЛокации(name string) VMTime {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return VMTime(time.Time(t).In(loc))
}

func (x VMTime) EvalBinOp(op VMOperation, y VMOperationer) (VMValuer, error) {
	switch op {
	case ADD:
		switch yy := y.(type) {
		case VMDurationer:
			return x.Добавить(yy.Duration()), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case SUB:
		switch yy := y.(type) {
		case VMDurationer:
			return x.Добавить(VMTimeDuration(-int64(yy.Duration()))), nil
		case VMTime:
			return x.Вычесть(yy), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case MUL:
		return VMNil, errors.New("Операция между значениями невозможна")
	case QUO:
		return VMNil, errors.New("Операция между значениями невозможна")
	case REM:
		return VMNil, errors.New("Операция между значениями невозможна")
	case EQL:
		switch yy := y.(type) {
		case VMDateTimer:
			return VMBool(x.Равно(yy.Time())), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case NEQ:
		switch yy := y.(type) {
		case VMDateTimer:
			return VMBool(!x.Равно(yy.Time())), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case GTR:
		switch yy := y.(type) {
		case VMDateTimer:
			return VMBool(x.Позже(yy.Time())), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case GEQ:
		switch yy := y.(type) {
		case VMDateTimer:
			return VMBool(x.Равно(yy.Time()) || x.Позже(yy.Time())), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case LSS:
		switch yy := y.(type) {
		case VMDateTimer:
			return VMBool(x.Раньше(yy.Time())), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case LEQ:
		switch yy := y.(type) {
		case VMDateTimer:
			return VMBool(x.Равно(yy.Time()) || x.Раньше(yy.Time())), nil
		}
		return VMNil, errors.New("Операция между значениями невозможна")
	case OR:
		return VMNil, errors.New("Операция между значениями невозможна")
	case LOR:
		return VMNil, errors.New("Операция между значениями невозможна")
	case AND:
		return VMNil, errors.New("Операция между значениями невозможна")
	case LAND:
		return VMNil, errors.New("Операция между значениями невозможна")
	case POW:
		return VMNil, errors.New("Операция между значениями невозможна")
	case SHR:
		return VMNil, errors.New("Операция между значениями невозможна")
	case SHL:
		return VMNil, errors.New("Операция между значениями невозможна")
	}
	return VMNil, errors.New("Неизвестная операция")
}

func (x VMTime) ConvertToType(nt reflect.Type) (VMValuer, error) {
	switch nt {
	case ReflectVMString:
		// сериализуем в json
		b, err := json.Marshal(x)
		if err != nil {
			return VMNil, err
		}
		return VMString(string(b)), nil
		// case ReflectVMInt:
	case ReflectVMTime:
		return x, nil
		// case ReflectVMBool:
		// case ReflectVMDecimal:
		// case ReflectVMSlice:
		// case ReflectVMStringMap:
	}

	return VMNil, errors.New("Приведение к типу невозможно")
}
