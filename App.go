package winclass

import (
    . "github.com/cwchiu/go-winapi"
    "syscall"
    "unsafe"
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
}

func (app *App) On(msg uint32, handler EventHandler){    
    app.EventMap[msg] = handler
}

func (app *App) Run(){    
    ShowWindow(app.HWnd, SW_NORMAL)
    UpdateWindow(app.HWnd)

    var msg MSG
    for GetMessage(&msg, HWND_TOP, 0, 0) == TRUE {
        TranslateMessage(&msg)
        DispatchMessage(&msg)
    }
}

func (app *App) Init(appName, title string) (error){
    app.AppName  = appName
    app.Title = title

    hCursor := LoadCursor(0, MAKEINTRESOURCE(IDC_ARROW))
    if hCursor == 0 {
        panic("LoadCursor")
    }
    
    szAppName := _T(appName)

    _default_wndproc := func (hwnd HWND, msg uint32, wParam, lParam uintptr) (result uintptr) {
        if app.EventMap[msg] != nil {
            return app.EventMap[msg](hwnd, msg, wParam, lParam)
        }
        return DefWindowProc(hwnd, msg, wParam, lParam)
    };

    var wc WNDCLASSEX
    wc.CbSize = uint32(unsafe.Sizeof(wc))
    wc.Style = CS_HREDRAW | CS_VREDRAW
    wc.LpfnWndProc = syscall.NewCallback(_default_wndproc)
    wc.HInstance = app.HInstance
    wc.HIcon = app.Icon
    wc.HCursor = hCursor
    wc.CbClsExtra = 0
    wc.CbWndExtra = 0
    wc.HbrBackground = app.BackgroundBrush
    wc.LpszMenuName = app.MenuName
    wc.LpszClassName = szAppName

    if atom := RegisterClassEx(&wc); atom == 0 {
        panic("RegisterClassEx")
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
        panic("CreateWindowEx")
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
    }
    
    return app, nil
}