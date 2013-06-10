package winclass

import (
    . "github.com/cwchiu/go-winapi"
    "syscall"
    "unsafe"
)

var _T func(s string) *uint16 = syscall.StringToUTF16Ptr

type EventHandler func(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr
type EvnatHandlerMap map[uint32]EventHandler
type App struct {
    AppName string
    Title string
    HInstance HINSTANCE
    HWnd HWND
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
    
    hInst := GetModuleHandle(nil)
    if hInst == 0 {
        panic("GetModuleHandle")
    }
    app.HInstance = hInst
    
    hIcon := LoadIcon(0, (*uint16)(unsafe.Pointer(uintptr(IDI_APPLICATION))))
    if hIcon == 0 {
        panic("LoadIcon")
    }

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
    wc.HInstance = hInst
    wc.HIcon = hIcon
    wc.HCursor = hCursor
    wc.CbClsExtra = 0
    wc.CbWndExtra = 0
    wc.HbrBackground = HBRUSH(GetStockObject(WHITE_BRUSH))
    wc.LpszMenuName = nil
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
        hInst,
        nil)

    if hWnd == 0 {
        panic("CreateWindowEx")
    }
    
    app.HWnd = hWnd
    
    return nil
}

func NewApp() (*App, error){
    app := &App{
        EventMap: make(EvnatHandlerMap),
    }
    
    return app, nil
}