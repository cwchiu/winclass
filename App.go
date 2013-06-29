package winclass

import (
    . "github.com/cwchiu/go-winapi"
    "syscall"
    "unsafe"
    "errors"    
    "container/list"
)

const (
    MSG_IGNORE = 9999
)

var _T func(s string) *uint16 = syscall.StringToUTF16Ptr

func Max(a, b int32) int32{
    if a>b {
        return a
    } else {
        return b
    }
}

func Min(a, b int32) int32{
    if a<b {
        return a
    } else {
        return b
    }
}
                  
type EventHandler func(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr
type EvnatHandlerMap map[uint32]EventHandler
type App struct {
    AppName string
    Title string
    HInstance HINSTANCE
    HWnd HWND
    Icon HICON    
    BackgroundBrush HBRUSH
    MenuName LPCTSTR
    EventMap EvnatHandlerMap
    ModelessDialogs list.List
}

func (app *App) On(msg uint32, handler EventHandler){    
    app.EventMap[msg] = handler
}

func (app *App) Run(){    
    ShowWindow(app.HWnd, SW_NORMAL)
    UpdateWindow(app.HWnd)

    var msg MSG
    for GetMessage(&msg, HWND_TOP, 0, 0) == TRUE {    
        if app.ModelessDialogs.Len() > 0 {
            for e := app.ModelessDialogs.Front(); e != nil; e = e.Next() {
                if IsDialogMessage (e.Value.(HWND), &msg) {
                    goto OutFor
                }
            }
        }        
        TranslateMessage(&msg)
        DispatchMessage(&msg)
OutFor: 
    }
}

func (app *App) AddModelessDialog(hwnd HWND) (error){
    app.ModelessDialogs.PushBack(hwnd)
    return nil
}

func (app *App) RegisterClass(className *uint16, wndproc EventHandler, menuName *uint16, hCursor HCURSOR, background HBRUSH) (ATOM, error){
    if hCursor == 0 {
        return 0, errors.New("Cursor Handler invalid")
    }
    
    var wc WNDCLASSEX
    wc.CbSize = uint32(unsafe.Sizeof(wc))
    wc.Style = CS_HREDRAW | CS_VREDRAW
    wc.LpfnWndProc = syscall.NewCallback(wndproc)
    wc.HInstance = app.HInstance
    wc.HIcon = app.Icon
    wc.HCursor = hCursor
    wc.CbClsExtra = 0
    wc.CbWndExtra = 0
    wc.HbrBackground = background
    wc.LpszMenuName = menuName
    wc.LpszClassName = className

    var atom ATOM
    if atom = RegisterClassEx(&wc); atom == 0 {
        return 0, errors.New("RegisterClassEx fail")
    }
    
    return atom, nil
}

func (app *App) Init(appName, title string) (error){
    app.AppName  = appName
    app.Title = title
    
    szAppName := _T(appName)   

    _, err := app.RegisterClass(szAppName, func (hwnd HWND, msg uint32, wParam, lParam uintptr) (result uintptr) {
        if app.EventMap[msg] != nil {
            ret := app.EventMap[msg](hwnd, msg, wParam, lParam)
            if ret != MSG_IGNORE {
                return ret
            }
        }
        return DefWindowProc(hwnd, msg, wParam, lParam)
    }, app.MenuName, LoadCursor(0, MAKEINTRESOURCE(IDC_ARROW)), app.BackgroundBrush)
    
    if err != nil {
        return err
    }
    
    hWnd := CreateWindowEx(
        0,
        szAppName,
        _T(title),
        WS_OVERLAPPEDWINDOW|WS_BORDER|WS_CAPTION|WS_SYSMENU|WS_MAXIMIZEBOX|WS_MINIMIZEBOX,
        CW_USEDEFAULT,
        CW_USEDEFAULT,
        CW_USEDEFAULT,
        CW_USEDEFAULT,
        HWND_TOP,
        0,
        app.HInstance,
        nil)

    if hWnd == 0 {
        return errors.New("CreateWindowEx fail")
    }
    
    app.HWnd = hWnd
    
    return nil
}

func NewApp() (*App, error){
    hInst := GetModuleHandle(nil)
    if hInst == 0 {
        panic("GetModuleHandle")
    }    
    
    app := &App{
        EventMap: make(EvnatHandlerMap),
        BackgroundBrush: HBRUSH(GetStockObject(WHITE_BRUSH)),
        Icon: LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(IDI_APPLICATION)))),
        HInstance: hInst,
        MenuName: nil,
        ModelessDialogs: list.List{},
    }
    
    return app, nil
}