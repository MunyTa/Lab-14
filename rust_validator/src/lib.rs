use std::ffi::CStr;
use std::os::raw::c_char;

#[no_mangle]
pub extern "C" fn validate_crypto_tick(
    symbol: *const c_char,
    price: f64,
    quantity: f64,
    timestamp_unix: i64,
) -> i32 {
    if symbol.is_null() {
        return 1;
    }

    let symbol = unsafe { CStr::from_ptr(symbol) };
    let symbol = match symbol.to_str() {
        Ok(value) => value.trim(),
        Err(_) => return 2,
    };

    if symbol.is_empty() {
        return 3;
    }
    if !symbol.chars().all(|ch| ch.is_ascii_uppercase() || ch.is_ascii_digit()) {
        return 4;
    }
    if !price.is_finite() || price <= 0.0 {
        return 5;
    }
    if !quantity.is_finite() || quantity <= 0.0 {
        return 6;
    }
    if timestamp_unix <= 0 {
        return 7;
    }

    0
}

